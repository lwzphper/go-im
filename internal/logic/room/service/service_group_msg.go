package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
)

// 群聊消息
func (s *Service) GroupMsg(n *connect.Node, data *room.Input) {
	if s.allServiceRoomMsg(n, data) {
		s.ack(n, data)
	}
}
