package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
)

// 创建房间
func (s *Service) Create(n *connect.Node, data *room.Input) {
	roomId := n.UserId // 房间id，使用用户id创建，为了简化判断逻辑。一个用户只能创建一个群聊

	roomName, ok := data.Data.(string)
	if !ok || roomName == "" {
		s.sendErrorMsg(n, data.RequestId, room.MethodCreateRoom, room.CodeValidateError, "房间名称格式有误")
	}

	// 房间已存在，返回错误信息
	isCreate, err := s.roomCache.Create(n.UserId, roomName)
	if err != nil {
		s.sendErrorMsg(n, data.RequestId, room.MethodCreateRoom, room.CodeError, "创建房间失败，请稍后再试。")
		return
	}
	if isCreate {
		s.sendErrorMsg(n, data.RequestId, room.MethodCreateRoom, room.CodeValidateError, "您已创建房间，不能重复创建")
		return
	}

	s.newRoom(roomId, roomName)

	roomInfo := room.RoomInfo{
		Id:   roomId,
		Name: roomName,
	}

	// 通知群用户，新创建了房间
	s.broadcastMsg(n, &room.Input{
		Data:   roomInfo,
		RoomId: roomId,
		Method: room.MethodCreateRoomNotice.Uint8(),
	})
}