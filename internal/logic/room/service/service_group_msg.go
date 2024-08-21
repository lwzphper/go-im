package service

import (
	types2 "go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

// 群聊消息
func (s *Service) groupMsg(n *types2.Node, data *types.Input) {
	if s.allServiceRoomMsg(n, data) {
		s.ack(n, data)
	}
}
