package service

import (
	"go-im/internal/connect"
)

// 处理关闭
func (s *Service) Close(n *connect.Node) {
	if n.RoomId > 0 {
		s.roomUserCache.Remove(n.RoomId, n.UserId)
		s.userServiceCache.Remove(n.RoomId, n.UserId)

		s.LeaveRoom(n, nil)
	}
}
