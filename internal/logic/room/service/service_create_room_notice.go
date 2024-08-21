package service

import (
	"go-im/internal/logic/room/types"
	types2 "go-im/internal/types"
)

// 新增房间通知
func (s *Service) createRoomNotice(n *types2.Node, data *types.Input) {
	n.DataQueue <- s.getOutput(n, data).QueueMsgData()
}
