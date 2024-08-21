package service

import (
	"go-im/internal/logic/room/types"
)

// 发送当前房间链接的消息
func (s *Service) SendRoomMsg(roomId uint64, data *types.QueueMsgData) {
	if r := s.getRoom(roomId); r != nil {
		s.pushRoom(r, data)
	}
}
