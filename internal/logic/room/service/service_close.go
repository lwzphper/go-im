package service

import (
	"go-im/internal/types"
)

// 处理关闭
func (s *Service) Close(n *types.Node) {
	if n.RoomId > 0 {
		s.roomUserCache.Remove(n.RoomId, n.UserId)
		s.userServiceCache.Remove(n.RoomId, n.UserId)

		s.leaveRoom(n, nil)
	}
}
