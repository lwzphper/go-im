package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
)

// 新增房间通知
func (s *Service) CreateRoomNotice(n *connect.Node, data *room.Input) {
	n.DataQueue <- s.getOutput(n, data).QueueMsgData()
}
