package event

import (
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
	"go-im/pkg/logger"
	"go.uber.org/zap"
)

const (
	ReadMsg               = "room:readMsg"      // 接收到客户端消息事件
	CloseConn             = "room:closeConn"    // 客户端连接关闭事件
	GatewayMsg            = "room:gatewayMsg"   // 网关广播事件
	ForceOfflineBroadcast = "room:forceOffline" // 强制下线事件
)

var RoomEvent = &roomEvent{
	//bus: event.NewAsyncEventBus(),
	bus: EventBus.New(),
}

type roomEvent struct {
	//bus *event.AsyncEventBus
	bus EventBus.Bus
}

// 订阅事件
func (r *roomEvent) Subscribe(event string, fn any) {
	if err := RoomEvent.bus.Subscribe(event, fn); err != nil {
		logger.Error("roomEvent Subscribe error", zap.Error(err))
	}
}

// 发布事件
func (r *roomEvent) Publish(event string, args ...any) {
	r.bus.Publish(event, args...)
}

// 强制下线广播事件
func (r *roomEvent) PushForceOfflineBroadcast(serviceId string, userId uint64) {
	r.bus.Publish(ForceOfflineBroadcast, serviceId, userId)
}

// 网关广播事件
func (r *roomEvent) PushGatewayMsg(wsConn *websocket.Conn, msg []byte) {
	r.bus.Publish(GatewayMsg, wsConn, msg)
}
