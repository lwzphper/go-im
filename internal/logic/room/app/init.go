package app

import (
	"github.com/gorilla/websocket"
	"go-im/internal/logic/room/service"
	"go-im/internal/pkg/event"
)

func Init() {
	srv := service.NewService()

	// 注册事件
	event.RoomEvent.Subscribe(event.EventReadMsg, func(userId uint64, msg []byte) {
		srv.Dispatch(userId, msg)
	})
	event.RoomEvent.Subscribe(event.EventGatewayMsg, func(wsConn *websocket.Conn, msg []byte) {
		srv.GatewayMsg(wsConn, msg)
	})
	event.RoomEvent.Subscribe(event.EventCloseConn, func(userId uint64) {
		srv.Close(userId)
	})
}
