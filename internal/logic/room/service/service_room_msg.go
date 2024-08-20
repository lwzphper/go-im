package service

import (
	"go-im/internal/logic/room"
)

// 发送当前房间链接的消息
func (s *Service) SendRoomMsg(roomId uint64, data *room.QueueMsgData) {
	if r := s.getRoom(roomId); r != nil {
		s.pushRoom(r, data)
	}
}
