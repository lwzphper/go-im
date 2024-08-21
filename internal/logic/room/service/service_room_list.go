package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go.uber.org/zap"
)

// 房间列表
func (s *Service) roomList(n *connect.Node, data *types.Input) {
	/*n := connect.GetNode(userId)
	if n == nil {
		return
	}*/
	list := s.roomCache.List()

	var result = types.RoomList{}
	for id, name := range list {
		roomId, err := util.StringToUint64(id)
		if err != nil {
			logger.Error("roomId parse error", zap.Error(err))
			continue
		}
		result = append(result, types.RoomInfo{
			Id:   roomId,
			Name: name,
		})
	}

	s.sendSuccessMsg(n, data.RequestId, types.MethodRoomList, result)
}
