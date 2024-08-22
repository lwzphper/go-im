package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

// 强制下线通知
func (s *Service) ForceOfflineBroadcast(serverId string, userId uint64) {
	data := types.QueueMsgData{
		Method:     types.MethodForceOfflineBroadcast,
		FromUid:    userId,
		FromServer: serverId,
	}
	connect.SendGatewayMsg(data.Marshal())
}
