package connect

import (
	"github.com/gorilla/websocket"
	"go-im/internal/event"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	MsgDefaultChannelSize = 1000 // 默认消息队列大小
	HeartBeatMaxErrorNum  = 2    // 心跳允许最大错误次数
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
	Mutex           sync.Mutex      // WS互斥锁
	Conn            *websocket.Conn // websocket连接
	UserId          uint64          // 用户ID
	RoomId          uint64          // 订阅的房间ID
	HeartbeatTime   int64           // 心跳时间
	HeartbeatErrNum uint8           // 心跳错误次数
	LoginTime       int64           // 登录时间
	DataQueue       chan []byte     // 消息队列
	BroadcastQueue  chan []byte     // 广播消息
	ServerAddr      string          // 服务器地址
	ServerId        string          // 服务器ID
	IsClose         bool            // 是否已关闭
}

func NewNode(conn *websocket.Conn, userId uint64, serverAddr, ServerId string, opts ...NodeOpt) *Node {
	nowTime := time.Now().Unix()
	node := &Node{
		Conn:           conn,
		UserId:         userId,
		HeartbeatTime:  nowTime,
		LoginTime:      nowTime,
		DataQueue:      make(chan []byte, MsgDefaultChannelSize),
		BroadcastQueue: make(chan []byte, MsgDefaultChannelSize),
		ServerAddr:     serverAddr,
		ServerId:       ServerId,
	}

	for _, opt := range opts {
		opt(node)
	}

	go node.handleRead()         // 读处理
	go node.handleWrite()        // 写处理
	go node.handleBroadcastMsg() // 处理广播消息

	return node
}

// 处理消息读取
func (n *Node) handleRead() {
	logger.Debugf("userId:%d 已连接", n.UserId)
	defer CloseConn(n)

	for {
		//_ = n.Conn.SetReadDeadline(time.Now().Add(writeWait))
		_, message, err := n.Conn.ReadMessage()
		/*if err != nil && WsErrorNeedClose(err) {
			return
		}*/
		if err != nil {
			logger.Debug("node 节点读取消息失败", zap.Error(err))
			return
		}

		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		logger.Debugf("接收到 userId:%d 数据：%s", n.UserId, string(message))

		event.RoomEvent.Publish(event.ReadMsg, n, message)
	}
}

// 处理消息写请求（给当前连接发送消息）
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
				_ = n.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			logger.Debugf("get data:%s", qData)

			if err := n.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("set write deadline error", zap.Error(err))
				return
			}

			if err := n.Conn.WriteMessage(websocket.TextMessage, qData); err != nil {
				logger.Error("write msg error", zap.Error(err))
				return
			}
		case <-ticker.C:
			logger.Debugf("用户id：%d 心跳检查", n.UserId)
			_ = n.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := n.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				n.HeartbeatErrNum++
				logger.Error("ping", zap.Error(err))
				// 心跳不通过，关闭连接
				if n.IsHeartbeatDeal() {
					logger.Info("heartbeat retry close", zap.String("用户id", util.Uint64ToString(n.UserId)))
					CloseConn(n)
					return
				}
			} else {
				n.HeartbeatTime = time.Now().Unix() // 更新心跳时间
			}
		}
	}
}

// 处理广播消息
func (n *Node) handleBroadcastMsg() {
	for {
		select {
		case wsData, ok := <-n.BroadcastQueue:
			if !ok {
				return
			}
			SendGatewayMsg(wsData)
		}
	}
}

// 检查是否连接是否存活
func (n *Node) IsHeartbeatDeal() bool {
	return n.HeartbeatErrNum >= HeartBeatMaxErrorNum
}

// GetAddr 获取ip地址
func (n *Node) GetAddr() string {
	return n.Conn.RemoteAddr().String()
}
