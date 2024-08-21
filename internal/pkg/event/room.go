package event

import (
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
)

const (
	EventReadMsg    = "room:readMsg"    // 接收到客户端消息事件
	EventCloseConn  = "room:closeConn"  // 客户端连接关闭事件
	EventGatewayMsg = "room:gatewayMsg" // 网关广播事件
)

var RoomEvent = &roomEvent{
	bus: EventBus.New(),
}

type roomEvent struct {
	bus EventBus.Bus
}

// 订阅事件
func (r *roomEvent) Subscribe(event string, fn any) {
	RoomEvent.bus.Subscribe(event, fn)
}

// 接收到客户端消息事件
func (r *roomEvent) PushReadMsg(userId uint64, msg []byte) {
	r.bus.Publish(EventReadMsg, userId, msg)
}

// 网关广播事件
func (r *roomEvent) PushGatewayMsg(wsConn *websocket.Conn, msg []byte) {
	r.bus.Publish(EventGatewayMsg, wsConn, msg)
}

// 客户端连接关闭事件
func (r *roomEvent) PushCloseConn(userId uint64) {
	r.bus.Publish(EventCloseConn, userId)
}
