package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

// 新增房间通知
func (s *Service) createRoomNotice(userId uint64, data *types.Input) {
	n := connect.GetNode(userId)
	if n == nil {
		return
	}
	n.DataQueue <- s.getOutput(n, data).Marshal()
}
