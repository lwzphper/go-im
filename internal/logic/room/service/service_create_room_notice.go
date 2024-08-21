package service

import (
	types2 "go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

// 新增房间通知
func (s *Service) createRoomNotice(n *types2.Node, data *types.Input) {
	n.DataQueue <- s.getOutput(n, data).Marshal()
}
