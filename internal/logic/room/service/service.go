package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
	"go-im/internal/logic/room/repo"
	"go-im/internal/logic/user/service"
	"sync"
)

var _ IService = (*Service)(nil)

func NewService() *Service {
	return &Service{
		userService:      service.NewUserService(),
		roomUserCache:    repo.NewRooUserCache(),
		userServiceCache: repo.NewUserServiceCache(),
		roomCache:        repo.NewRoomCache(),
		roomsManager:     make(map[uint64]*Room),
	}
}

type IService interface {
	// 创建房间
	Create(n *connect.Node, data *room.Input)
	// 新增房间通知
	CreateRoomNotice(n *connect.Node, data *room.Input)
	// 群聊消息
	GroupMsg(n *connect.Node, data *room.Input)
	// 加入房间
	Join(n *connect.Node, data *room.Input)
	// 离开房间
	LeaveRoom(n *connect.Node, data *room.Input)
	// 房间列表
	RoomList(n *connect.Node, data *room.Input)
	// 获取房间用户列表
	UserList(n *connect.Node, data *room.Input)
	// 一对一消息
	Normal(n *connect.Node, data *room.Input)
	// 发送当前房间链接的消息
	ServerRoomMsg(n *connect.Node, data *room.Input)
	// 关闭操作
	Close(n *connect.Node)
}

type Service struct {
	userService      service.IService
	userServiceCache *repo.UserServiceCache
	roomUserCache    *repo.RoomUserCache
	roomCache        *repo.RoomCache
	roomsManager     map[uint64]*Room
	newRoomLock      sync.Mutex
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
func (s *Service) sendSuccessMsg(n *connect.Node, RequestId string, method room.MsgMethod, data any) {
	n.DataQueue <- &room.QueueMsgData{
		RequestId: RequestId,
		Method:    method,
		Data:      data,
		Code:      room.CodeSuccess,
	}
}

// 发送错误信息
func (s *Service) sendErrorMsg(n *connect.Node, RequestId string, method room.MsgMethod, code room.Code, Msg string) {
	n.DataQueue <- &room.QueueMsgData{
		RequestId: RequestId,
		Method:    method,
		Code:      code,
		Msg:       Msg,
	}
}

// 获取输出结果
func (s *Service) getOutput(n *connect.Node, data *room.Input) *room.Output {
	return &room.Output{
		RequestId:    data.RequestId,
		Code:         room.CodeSuccess,
		Method:       room.MsgMethod(data.Method),
		Data:         data.Data,
		RoomId:       data.RoomId,
		FromUid:      n.UserId,
		FromUsername: s.userService.UserIdName(n.UserId),
		ToUid:        data.ToUid,
		FromServer:   n.ServerId,
	}
}

// 发送房间消息（全部服务器）
func (s *Service) allServiceRoomMsg(n *connect.Node, data *room.Input) bool {
	if !s.isInRoom(n, data.RoomId) {
		s.sendErrorMsg(n, data.RequestId, room.MethodGroup, room.CodeValidateError, "未加入群聊")
		return false
	}

	// 广播消息
	n.BroadcastQueue <- s.getOutput(n, data)

	// 推送当前服务指定房间的全部用户
	s.sendServerRoom(n, data)

	return true
}

// ack 确认
func (s *Service) ack(n *connect.Node, data *room.Input) {
	s.sendSuccessMsg(n, data.RequestId, room.MethodServiceAck, nil)
}

// 发送房间消息（当前服务器）
func (s *Service) sendServerRoom(n *connect.Node, data *room.Input) {
	out := s.getOutput(n, data)

	if r := s.getRoom(n.RoomId); r != nil {
		s.pushRoom(r, out.QueueMsgData())
	}
}

// 广播消息（全部在线用户，区分房间）
func (s *Service) broadcastMsg(n *connect.Node, data *room.Input) {
	// 广播消息
	out := s.getOutput(n, data)

	n.BroadcastQueue <- out

	// 推送当前服务指定房间的全部用户
	connect.PushAll(out.QueueMsgData())
}
