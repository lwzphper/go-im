package service

import (
	"github.com/gorilla/websocket"
	"go-im/internal/connect"
	"go-im/internal/logic/room/repo"
	"go-im/internal/logic/room/types"
	"go-im/internal/logic/user/service"
	"sync"
)

var _ IService = (*Service)(nil)

func NewService() *Service {
	srv := &Service{
		userService:      service.NewUserService(),
		roomUserCache:    repo.NewRooUserCache(),
		userServiceCache: repo.NewUserServiceCache(),
		roomCache:        repo.NewRoomCache(),
		roomsManager:     make(map[uint64]*Room),
		strategy:         MsgStrategy{},
	}

	srv.strategy.Register(types.MethodCreateRoom, srv.create)
	srv.strategy.Register(types.MethodJoinRoom, srv.join)
	srv.strategy.Register(types.MethodRoomUser, srv.userList)
	srv.strategy.Register(types.MethodNormal, srv.normal)
	srv.strategy.Register(types.MethodGroup, srv.groupMsg)
	srv.strategy.Register(types.MethodRoomList, srv.roomList)
	srv.strategy.Register(types.MethodCreateRoomNotice, srv.createRoomNotice)
	srv.strategy.Register(types.MethodOffline, srv.leaveRoom)

	return srv
}

type IService interface {
	// 分发消息
	Dispatch(userId uint64, message []byte)
	// 网关消息
	GatewayMsg(wsConn *websocket.Conn, message []byte)
	// 发送当前房间链接的消息
	SendRoomMsg(roomId uint64, data *types.QueueMsgData)
	// 关闭操作
	Close(userId uint64)
	// 强制下线
	ForceOfflineBroadcast(serverId string, userId uint64)
}

type Service struct {
	userService      service.IService
	userServiceCache *repo.UserServiceCache
	roomUserCache    *repo.RoomUserCache
	roomCache        *repo.RoomCache
	roomsManager     map[uint64]*Room
	newRoomLock      sync.Mutex
	strategy         MsgStrategy
}

type Room struct {
	RoomId    uint64 // 房间ID
	name      string // 房间名称
	clients   map[uint64]*connect.Node
	joinLock  sync.RWMutex
	leaveLock sync.RWMutex
	pushLock  sync.RWMutex
}

// 判断用户是否在加入房间
func (s *Service) isInRoom(n *connect.Node, roomId uint64) bool {
	if roomId == 0 {
		return false
	}

	if n.RoomId == roomId {
		return true
	}

	if n.RoomId == 0 {
		if s.roomUserCache.Exists(n.UserId, roomId) {
			n.RoomId = roomId
			return true
		}
		return false
	}

	return false
}

// 发送成功消息
func (s *Service) sendSuccessMsg(n *connect.Node, RequestId string, method types.MsgMethod, data any) {
	result := types.Output{
		RequestId: RequestId,
		Code:      types.CodeSuccess,
		Msg:       types.CodeSuccess.Name(),
		Method:    method,
		Data:      data,
		RoomId:    n.RoomId,
	}

	n.DataQueue <- result.Marshal()
}

// 发送错误信息
func (s *Service) sendErrorMsg(n *connect.Node, RequestId string, method types.MsgMethod, code types.Code, msg string) {
	if msg == "" {
		msg = code.Name()
	}
	result := types.Output{
		RequestId: RequestId,
		Code:      code,
		Msg:       msg,
		Method:    method,
		RoomId:    n.RoomId,
	}

	n.DataQueue <- result.Marshal()
}

// 获取输出结果
func (s *Service) getOutput(n *connect.Node, data *types.Input) *types.Output {
	return &types.Output{
		RequestId:    data.RequestId,
		Code:         types.CodeSuccess,
		Method:       types.MsgMethod(data.Method),
		Data:         data.Data,
		RoomId:       data.RoomId,
		FromUid:      n.UserId,
		FromUsername: s.userService.UserIdName(n.UserId),
		ToUid:        data.ToUid,
		FromServer:   n.ServerId,
	}
}

// 发送房间消息（全部服务器）
func (s *Service) allServiceRoomMsg(n *connect.Node, data *types.Input) bool {
	if !s.isInRoom(n, data.RoomId) {
		s.sendErrorMsg(n, data.RequestId, types.MethodGroup, types.CodeValidateError, "未加入群聊")
		return false
	}

	// 广播消息
	n.BroadcastQueue <- s.getOutput(n, data).QueueMsgData().Marshal()

	// 推送当前服务指定房间的全部用户
	s.sendServerRoom(n, data)

	return true
}

// ack 确认
func (s *Service) ack(n *connect.Node, data *types.Input) {
	s.sendSuccessMsg(n, data.RequestId, types.MethodServiceAck, nil)
}

// 发送房间消息（当前服务器）
func (s *Service) sendServerRoom(n *connect.Node, data *types.Input) {
	out := s.getOutput(n, data)

	if r := s.getRoom(n.RoomId); r != nil {
		s.pushRoom(r, out.QueueMsgData())
	}
}

// 广播消息（全部在线用户，区分房间）
func (s *Service) broadcastMsg(n *connect.Node, data *types.Input) {
	// 广播消息
	out := s.getOutput(n, data)

	n.BroadcastQueue <- out.QueueMsgData().Marshal()

	// 推送当前服务指定房间的全部用户
	connect.PushAll(out.QueueMsgData())
}
