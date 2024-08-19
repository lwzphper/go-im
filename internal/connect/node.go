package connect

import (
	"errors"
	"github.com/gorilla/websocket"
	"go-im/internal/connect/repo"
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
	mutex           sync.Mutex         // WS互斥锁
	conn            *websocket.Conn    // websocket连接
	UserId          uint64             // 用户ID
	RoomId          uint64             // 订阅的房间ID
	heartbeatTime   int64              // 心跳时间
	heartbeatErrNum uint8              // 心跳错误次数
	LoginTime       int64              // 登录时间
	DataQueue       chan *QueueMsgData // 消息队列
	broadcastQueue  chan *Output       // 广播消息
	ServerAddr      string             // 服务器地址
	ServerId        string             // 服务器ID
	IsClose         bool               // 是否已关闭
}

func NewNode(conn *websocket.Conn, userId uint64, serverAddr, ServerId string, opts ...NodeOpt) *Node {
	nowTime := time.Now().Unix()
	node := &Node{
		conn:           conn,
		UserId:         userId,
		heartbeatTime:  nowTime,
		LoginTime:      nowTime,
		DataQueue:      make(chan *QueueMsgData, MsgDefaultChannelSize),
		broadcastQueue: make(chan *Output, MsgDefaultChannelSize),
		ServerAddr:     serverAddr,
		ServerId:       ServerId,
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

		data, err := UnMarshalInput(message)
		if err != nil {
			logger.Infof("用户：%d 消息格式有误：%s", n.UserId, string(message))
			n.sendErrorMsg("", MethodServiceNotice, CodeValidateError, "消息格式有误")
			continue
		}

		// 设置房间id默认值
		if data.RoomId == 0 {
			data.RoomId = n.RoomId
		}

		switch data.Method {
		case MethodCreateRoom.Uint8(): // 创建房间
			n.handleCreateRoom(data)
		case MethodJoinRoom.Uint8(): // 加入房间
			n.handleJoinRoom(data)
		case MethodRoomUser.Uint8(): // 获取房间用户
			n.handleRoomUser(data)
		case MethodNormal.Uint8(): // 普通消息。发送指定用户
			n.handleNormalMsg(data)
		case MethodGroup.Uint8(): // 群聊消息
			if n.RoomGroupMsg(data) {
				n.sendSuccessMsg(data.RequestId, MethodServiceAck, nil)
			}
		case MethodRoomList.Uint8(): // 房间列表
			n.handleRoomList(data.RequestId)
		case MethodCreateRoomNotice.Uint8(): // 新增房间通知
			n.DataQueue <- n.getOutput(data).QueueMsgData()
		case MethodOffline.Uint8(): // 下线通知
			n.handleLeaveRoom()
		}
	}
}

// 判断用户是否在加入房间
func (n *Node) isInRoom(roomId uint64) bool {
	if roomId == 0 {
		return false
	}

	if n.RoomId == roomId {
		return true
	}

	if n.RoomId == 0 {
		if repo.RoomUserCache.Exists(n.UserId, roomId) {
			n.RoomId = roomId
			return true
		}
		return false
	}

	return false
}

// 获取房间用户
func (n *Node) handleRoomUser(data *Input) {
	if !n.isInRoom(data.RoomId) {
		n.sendErrorMsg(data.RequestId, MethodRoomUser, CodeValidateError, "请选择房间或群组")
		return
	}

	list := GetRoom(n.RoomId).getUserList()
	n.sendSuccessMsg(data.RequestId, MethodRoomUser, list)
}

// 加入房间
func (n *Node) handleJoinRoom(data *Input) {
	if data.RoomId == 0 {
		n.sendErrorMsg(data.RequestId, MethodJoinRoom, CodeValidateError, "请选择房间或群组")
		return
	}

	room := GetRoom(data.RoomId)
	if room == nil {
		n.sendErrorMsg(data.RequestId, MethodJoinRoom, CodeValidateError, "房间不存在")
		return
	}

	// 获取用户名称
	username := UserManger.name(n.UserId)
	if username == "" {
		n.sendErrorMsg(data.RequestId, MethodJoinRoom, CodeValidateError, "获取用户信息失败，请稍后再试。")
		return
	}
	// 加入房间
	room.Join(n, username)

	// 通知群用户
	n.RoomGroupMsg(&Input{
		Data: UserItem{
			Id:   n.UserId,
			Name: username,
		},
		RoomId: data.RoomId,
		Method: MethodOnline.Uint8(),
	})

	n.sendSuccessMsg(data.RequestId, MethodJoinRoom, &RoomInfo{
		Id:   room.RoomId,
		Name: room.name,
	})
}

// 创建房间
func (n *Node) handleCreateRoom(data *Input) {
	roomId := n.UserId // 房间id，使用用户id创建，为了简化判断逻辑。一个用户只能创建一个群聊

	roomName, ok := data.Data.(string)
	if !ok || roomName == "" {
		n.sendErrorMsg(data.RequestId, MethodCreateRoom, CodeValidateError, "房间名称格式有误")
	}

	// 房间已存在，返回错误信息
	isCreate, err := repo.RoomCache.Create(n.UserId, roomName)
	if err != nil {
		n.sendErrorMsg(data.RequestId, MethodCreateRoom, CodeError, "创建房间失败，请稍后再试。")
		return
	}
	if isCreate {
		n.sendErrorMsg(data.RequestId, MethodCreateRoom, CodeValidateError, "您已创建房间，不能重复创建")
		return
	}

	newRoom(roomId, roomName)

	roomInfo := RoomInfo{
		Id:   roomId,
		Name: roomName,
	}

	// 通知群用户，新创建了房间
	n.broadcastMsg(&Input{
		Data:   roomInfo,
		RoomId: roomId,
		Method: MethodCreateRoomNotice.Uint8(),
	})
}

// 房间列表
func (n *Node) handleRoomList(requestId string) {
	list := repo.RoomCache.List()

	var result = RoomList{}
	for id, name := range list {
		roomId, err := util.StringToUint64(id)
		if err != nil {
			logger.Error("roomId parse error", zap.Error(err))
			continue
		}
		result = append(result, RoomInfo{
			Id:   roomId,
			Name: name,
		})
	}

	n.sendSuccessMsg(requestId, MethodRoomList, result)
}

// 处理普通消息
func (n *Node) handleNormalMsg(data *Input) {
	if data.ToUid == 0 {
		n.sendErrorMsg(data.RequestId, MethodNormal, CodeValidateError, "未选择发送的用户")
		return
	}

	n.sendErrorMsg(data.RequestId, MethodNormal, CodeError, "目前只支持群聊，暂不支持私聊")
	return
}

// 发送当前服务器的房间用户
func (n *Node) sendServerRoom(data *Input) {
	out := n.getOutput(data)

	if room := GetRoom(n.RoomId); room != nil {
		room.Push(out.QueueMsgData())
	}
}

// 获取输出结果
func (n *Node) getOutput(data *Input) *Output {
	return &Output{
		RequestId:    data.RequestId,
		Code:         CodeSuccess,
		Method:       MsgMethod(data.Method),
		Data:         data.Data,
		RoomId:       data.RoomId,
		FromUid:      n.UserId,
		FromUsername: UserManger.name(n.UserId),
		ToUid:        data.ToUid,
		FromServer:   n.ServerId,
	}
}

// 广播消息（全部在线用户，区分房间）
func (n *Node) broadcastMsg(data *Input) {
	// 广播消息
	out := n.getOutput(data)

	n.broadcastQueue <- out

	// 推送当前服务指定房间的全部用户
	pushAll(out.QueueMsgData())
}

// 退出房间
func (n *Node) handleLeaveRoom() {
	// 通知房间其他用户下线（不能通过 chan 通知，因为上面已将相关 chan 关闭）
	n.offlineNotify()

	// 从房间中删除连接（这里会将链接的 roomId 重置为0，因此要放到最后）
	if room := GetRoom(n.RoomId); room != nil {
		room.Leave(n)
	}
}

// RoomGroupMsg 群聊消息
func (n *Node) RoomGroupMsg(data *Input) bool {
	if !n.isInRoom(data.RoomId) {
		n.sendErrorMsg(data.RequestId, MethodGroup, CodeValidateError, "未加入群聊")
		return false
	}

	// 广播消息
	n.broadcastQueue <- n.getOutput(data)

	// 推送当前服务指定房间的全部用户
	n.sendServerRoom(data)

	return true
}

// 发送错误信息
func (n *Node) sendErrorMsg(RequestId string, method MsgMethod, code Code, Msg string) {
	n.DataQueue <- &QueueMsgData{
		RequestId: RequestId,
		Method:    method,
		Code:      code,
		Msg:       Msg,
	}
}

// 发送成功消息
func (n *Node) sendSuccessMsg(RequestId string, method MsgMethod, data any) {
	n.DataQueue <- &QueueMsgData{
		RequestId: RequestId,
		Method:    method,
		Data:      data,
		Code:      CodeSuccess,
	}
}

// 转发消息（广播给其他节点，需要排除当前节点）
func (n *Node) handleBroadcastMsg() {
	var ws *websocket.Conn
	for {
		select {
		case wsData, ok := <-n.broadcastQueue:
			if !ok {
				return
			}

			if ws = getGatewayClient(); ws != nil {
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

			data := Output{
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
	close(n.broadcastQueue)

	// 删除用户连接映射
	DeleteNode(n.UserId)

	if n.RoomId > 0 {
		repo.RoomUserCache.Remove(n.RoomId, n.UserId)
		repo.UserServiceCache.Remove(n.RoomId, n.UserId)

		n.handleLeaveRoom()
	}

	logger.Debugf("用户：%d 程序已关闭连接", n.UserId)
}

// 下线广播（不能通过 chan 通知，因为关闭客户端时已将相关 chan 关闭）
func (n *Node) offlineNotify() {
	name := UserManger.name(n.UserId)
	data := Output{
		Method: MethodOffline,
		Data: UserItem{
			Id:   n.UserId,
			Name: name,
		},
		RoomId:     n.RoomId,
		FromServer: n.ServerId,
	}

	// 广播通知其他服务
	if ws := getGatewayClient(); ws != nil {
		_ = ws.WriteMessage(websocket.TextMessage, data.Marshal())
	}

	// 推送当前服务指定房间的全部用户
	if room := GetRoom(n.RoomId); room != nil {
		room.Push(data.QueueMsgData())
	}
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
