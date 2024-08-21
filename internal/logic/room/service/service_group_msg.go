package service

import (
	"go-im/internal/logic/room/types"
	types2 "go-im/internal/types"
)

// 群聊消息
func (s *Service) groupMsg(n *types2.Node, data *types.Input) {
	if s.allServiceRoomMsg(n, data) {
		s.ack(n, data)
	}
}
