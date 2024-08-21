package service

import (
	"github.com/gorilla/websocket"
	"go-im/internal/connect"
	roomType "go-im/internal/logic/room/types"
	"go-im/internal/types"
)

// 离开房间
func (s *Service) leaveRoom(n *types.Node, data *roomType.Input) {
	// 下线广播
	s.offlineNotify(n)

	// 从房间中删除连接（这里会将链接的 roomId 重置为0，因此要放到最后）
	if room := s.getRoom(n.RoomId); room != nil {
		s.handleLeaveRoom(room, n)
	}
}

// 下线广播（不能通过 chan 通知，因为关闭客户端时已将相关 chan 关闭）
func (s *Service) offlineNotify(n *types.Node) {
	name := s.userService.UserIdName(n.UserId)
	data := roomType.Output{
		Method: roomType.MethodOffline,
		Data: roomType.UserItem{
			Id:   n.UserId,
			Name: name,
		},
		RoomId:     n.RoomId,
		FromServer: n.ServerId,
	}

	// 广播通知其他服务
	if ws := connect.GetGatewayClient(); ws != nil {
		_ = ws.WriteMessage(websocket.TextMessage, data.Marshal())
	}

	// 推送当前服务指定房间的全部用户
	if room := s.getRoom(n.RoomId); room != nil {
		s.pushRoom(room, data.QueueMsgData())
	}
}
