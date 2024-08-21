package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

// 群聊消息
func (s *Service) groupMsg(n *connect.Node, data *types.Input) {
	/*n := connect.GetNode(userId)
	if n == nil {
		return
	}*/
	if s.allServiceRoomMsg(n, data) {
		s.ack(n, data)
	}
}
