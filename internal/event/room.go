package event

import (
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
	types2 "go-im/internal/connect"
	"go-im/internal/logic/room/service"
)

const (
	EventReadMsg    = "room:readMsg"    // 接收到客户端消息事件
	EventCloseConn  = "room:closeConn"  // 客户端连接关闭事件
	EventGatewayMsg = "room:gatewayMsg" // 网关广播事件
)

var RoomEvent = &roomEvent{
	bus:     EventBus.New(),
	roomSrv: service.NewService(),
}

type roomEvent struct {
	bus     EventBus.Bus
	roomSrv service.IService
}

// 接收到客户端消息事件
func (r *roomEvent) PushReadMsg(n *types2.Node, msg []byte) {
	r.bus.Publish(EventReadMsg, n, msg)
}

// 网关广播事件
func (r *roomEvent) PushGatewayMsg(wsConn *websocket.Conn, msg []byte) {
	r.bus.Publish(EventGatewayMsg, wsConn, msg)
}

// 客户端连接关闭事件
func (r *roomEvent) PushCloseConn(n *types2.Node) {
	r.bus.Publish(EventCloseConn, n)
}

// 注册事件
func init() {
	RoomEvent.bus.Subscribe(EventReadMsg, func(n *types2.Node, msg []byte) {
		RoomEvent.roomSrv.Dispatch(n, msg)
	})
	RoomEvent.bus.Subscribe(EventGatewayMsg, func(wsConn *websocket.Conn, msg []byte) {
		RoomEvent.roomSrv.GatewayMsg(wsConn, msg)
	})
	RoomEvent.bus.Subscribe(EventCloseConn, func(n *types2.Node) {
		RoomEvent.roomSrv.Close(n)
	})
}
