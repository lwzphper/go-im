package app

import (
	"go-im/internal/event"
	"go-im/internal/logic/room/service"
)

func Init() {
	srv := service.NewService()

	// 注册事件
	event.RoomEvent.SubscribeAsync(event.ReadMsg, srv.Dispatch)
	event.RoomEvent.SubscribeAsync(event.GatewayMsg, srv.GatewayMsg)
	event.RoomEvent.SubscribeAsync(event.CloseConn, srv.Close)
}
