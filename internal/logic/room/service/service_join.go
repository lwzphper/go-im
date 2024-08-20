package service

import (
	"go-im/internal/connect"
	roomType "go-im/internal/logic/room"
)

// 加入房间
func (s *Service) Join(n *connect.Node, data *roomType.Input) {
	if data.RoomId == 0 {
		s.sendErrorMsg(n, data.RequestId, roomType.MethodJoinRoom, roomType.CodeValidateError, "请选择房间或群组")
		return
	}

	room := s.getRoom(data.RoomId)
	if room == nil {
		s.sendErrorMsg(n, data.RequestId, roomType.MethodJoinRoom, roomType.CodeValidateError, "房间不存在")
		return
	}

	// 获取用户名称
	username := s.userService.UserIdName(n.UserId)
	if username == "" {
		s.sendErrorMsg(n, data.RequestId, roomType.MethodJoinRoom, roomType.CodeValidateError, "获取用户信息失败，请稍后再试。")
		return
	}
	// 加入房间
	s.joinRoom(room, n, username)

	// 通知群用户
	s.allServiceRoomMsg(n, &roomType.Input{
		Data: roomType.UserItem{
			Id:   n.UserId,
			Name: username,
		},
		RoomId: data.RoomId,
		Method: roomType.MethodOnline.Uint8(),
	})

	s.sendSuccessMsg(n, data.RequestId, roomType.MethodJoinRoom, &roomType.RoomInfo{
		Id:   room.RoomId,
		Name: room.name,
	})
}
