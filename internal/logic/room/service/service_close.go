package service

import (
	"go-im/internal/connect"
)

// 处理关闭
func (s *Service) Close(userId uint64) {
	n := connect.GetNode(userId)
	if n == nil {
		return
	}

	if n.RoomId > 0 {
		s.roomUserCache.Remove(n.RoomId, n.UserId)
		s.userServiceCache.Remove(n.RoomId, n.UserId)

		s.leaveRoom(n.UserId, nil)
	}
}
