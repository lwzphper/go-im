package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
)

// 发送当前房间链接的消息
func (s *Service) ServerRoomMsg(n *connect.Node, data *room.Input) {
	s.sendServerRoom(n, data)
}
