package service

import (
	"go-im/internal/connect"
	roomType "go-im/internal/logic/room/types"
)

// 获取房间用户列表
func (s *Service) userList(n *connect.Node, data *roomType.Input) {
	if !s.isInRoom(n, data.RoomId) {
		s.sendErrorMsg(n, data.RequestId, roomType.MethodRoomUser, roomType.CodeValidateError, "请选择房间或群组")
		return
	}

	list := s.roomUserList(n.RoomId)
	s.sendSuccessMsg(n, data.RequestId, roomType.MethodRoomUser, list)
}
