package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go.uber.org/zap"
)

// 房间列表
func (s *Service) RoomList(n *connect.Node, data *room.Input) {
	list := s.roomCache.List()

	var result = room.RoomList{}
	for id, name := range list {
		roomId, err := util.StringToUint64(id)
		if err != nil {
			logger.Error("roomId parse error", zap.Error(err))
			continue
		}
		result = append(result, room.RoomInfo{
			Id:   roomId,
			Name: name,
		})
	}

	s.sendSuccessMsg(n, data.RequestId, room.MethodRoomList, result)
}
