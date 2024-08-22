package app

import (
	"go-im/internal/event"
	"go-im/internal/logic/room/service"
)

func Init() {
	srv := service.NewService()

	// 注册事件
	event.RoomEvent.Subscribe(event.ReadMsg, srv.Dispatch)
	event.RoomEvent.Subscribe(event.GatewayMsg, srv.GatewayMsg)
	event.RoomEvent.Subscribe(event.CloseConn, srv.Close)
	event.RoomEvent.Subscribe(event.ForceOfflineBroadcast, srv.ForceOfflineBroadcast)
}
