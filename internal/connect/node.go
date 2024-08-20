package connect

import (
	"errors"
	"github.com/gorilla/websocket"
	"go-im/internal/logic/room"
	"go-im/internal/logic/room/app"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go.uber.org/zap"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	MsgDefaultChannelSize = 1000 // 默认消息队列大小
	pongWait              = 60 * time.Second
	writeWait             = 10 * time.Second // 写超时时间
	pingPeriod            = 60 * time.Second // 心跳时间周期
	HeartBeatMaxErrorNum  = 2                // 心跳允许最大错误次数
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type NodeOpt func(node *Node)

func WithNodeLoginTime(t int64) NodeOpt {
	return func(node *Node) {
		node.LoginTime = t
	}
}

type Node struct {
	mutex           sync.Mutex              // WS互斥锁
	conn            *websocket.Conn         // websocket连接
	UserId          uint64                  // 用户ID
	RoomId          uint64                  // 订阅的房间ID
	heartbeatTime   int64                   // 心跳时间
	heartbeatErrNum uint8                   // 心跳错误次数
	LoginTime       int64                   // 登录时间
	DataQueue       chan *room.QueueMsgData // 消息队列
	BroadcastQueue  chan *room.Output       // 广播消息
	ServerAddr      string                  // 服务器地址
	ServerId        string                  // 服务器ID
	IsClose         bool                    // 是否已关闭

	roomApp *app.RoomApp
}

func NewNode(conn *websocket.Conn, userId uint64, serverAddr, ServerId string, roomApp *app.RoomApp, opts ...NodeOpt) *Node {
	nowTime := time.Now().Unix()
	node := &Node{
		conn:           conn,
		UserId:         userId,
		heartbeatTime:  nowTime,
		LoginTime:      nowTime,
		DataQueue:      make(chan *room.QueueMsgData, MsgDefaultChannelSize),
		BroadcastQueue: make(chan *room.Output, MsgDefaultChannelSize),
		ServerAddr:     serverAddr,
		ServerId:       ServerId,
		roomApp:        roomApp,
	}

	for _, opt := range opts {
		opt(node)
	}

	go node.handleRead()         // 读处理
	go node.handleWrite()        // 写处理
	go node.handleBroadcastMsg() // 处理广播消息

	// 用户跟节点的映射
	SetNode(node.UserId, node)

	return node
}

// handleRead 处理消息读取
func (n *Node) handleRead() {
	logger.Debugf("userId:%d 已连接", n.UserId)
	defer n.Close()

	for {
		//_ = n.conn.SetReadDeadline(time.Now().Add(writeWait))
		_, message, err := n.conn.ReadMessage()
		/*if err != nil && WsErrorNeedClose(err) {
			return
		}*/
		if err != nil {
			logger.Debug("node 节点读取消息失败", zap.Error(err))
			return
		}

		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		logger.Debugf("接收到 userId:%d 数据：%s", n.UserId, string(message))

		n.roomApp.Dispatch(n, message)
	}
}

// 转发消息（广播给其他节点，需要排除当前节点）
func (n *Node) handleBroadcastMsg() {
	var ws *websocket.Conn
	for {
		select {
		case wsData, ok := <-n.BroadcastQueue:
			if !ok {
				return
			}

			if ws = GetGatewayClient(); ws != nil {
				msg := wsData.QueueMsgData().Marshal()
				logger.Debug("发送广播消息：" + string(msg))
				err := ws.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					logger.Debug("发送广播消息失败：" + err.Error())
				}
			}
		}
	}
}

// handleWrite 处理消息写请求（给当前连接发送消息）
func (n *Node) handleWrite() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		//n.Close()
	}()

	for {
		select {
		case qData, ok := <-n.DataQueue:
			if !ok {
				// 连接关闭
				_ = n.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			logger.Debugf("[conn %d] get data from queue:%s", n.UserId, qData.Data)

			if err := n.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("set write deadline error", zap.Error(err))
				return
			}

			data := room.Output{
				RequestId:    qData.RequestId,
				Code:         qData.Code,
				Msg:          qData.Msg,
				Method:       qData.Method,
				Data:         qData.Data,
				FromUid:      qData.FromUid,
				FromUsername: qData.FromUsername,
				RoomId:       n.RoomId,
				FromServer:   qData.FromServer,
			}

			if data.Msg == "" {
				data.Msg = data.Code.Name()
			}

			if err := n.conn.WriteMessage(websocket.TextMessage, data.Marshal()); err != nil {
				logger.Error("write msg error", zap.Error(err))
				return
			}
		case <-ticker.C:
			logger.Debugf("用户id：%d 心跳检查", n.UserId)
			_ = n.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := n.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				n.heartbeatErrNum++
				logger.Error("ping", zap.Error(err))
				// 心跳不通过，关闭连接
				if n.isHeartbeatDeal() {
					logger.Info("heartbeat retry close", zap.String("用户id", util.Uint64ToString(n.UserId)))
					n.Close()
					return
				}
			} else {
				n.heartbeatTime = time.Now().Unix() // 更新心跳时间
			}
		}
	}
}

// Close 关闭连接
func (n *Node) Close() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.IsClose {
		return
	}

	n.IsClose = true

	if err := n.conn.Close(); err != nil {
		logger.Error("close node connect error", zap.Error(err))
	}
	close(n.DataQueue)
	close(n.BroadcastQueue)

	// 删除用户连接映射
	DeleteNode(n.UserId)

	n.roomApp.Close(n)

	logger.Debugf("用户：%d 程序已关闭连接", n.UserId)
}

// 读取conn错误。返回结果：是否需要终止监听
func (n *Node) handleReadErr(err error) bool {
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) {
		logger.Debug("连接关闭", zap.String("用户id", util.Uint64ToString(n.UserId)))
		n.Close()
		return true
	}

	str := err.Error()

	// 服务器主动关闭连接
	if strings.HasSuffix(str, "use of closed network connection") {
		n.Close()
		return true
	}

	if err == io.EOF {
		n.Close()
		return true
	}

	// SetReadDeadline 之后，超时返回的错误
	if strings.HasSuffix(str, "i/o timeout") {
		n.Close()
		return true
	}

	logger.Debug("read tcp error：", zap.Uint64("user_id", n.UserId), zap.Error(err))
	return false
}

// 检查是否连接是否存活
func (n *Node) isHeartbeatDeal() bool {
	return n.heartbeatErrNum >= HeartBeatMaxErrorNum
}

// GetAddr 获取ip地址
func (n *Node) GetAddr() string {
	return n.conn.RemoteAddr().String()
}
