package event

import (
	"github.com/asaskevich/EventBus"
	"go-im/pkg/logger"
	"go.uber.org/zap"
)

const (
	ReadMsg    = "room:readMsg"    // 接收到客户端消息事件
	CloseConn  = "room:closeConn"  // 客户端连接关闭事件
	GatewayMsg = "room:gatewayMsg" // 网关广播事件
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
func (r *roomEvent) SubscribeAsync(event string, fn any) {
	if err := RoomEvent.bus.SubscribeAsync(event, fn, false); err != nil {
		logger.Error("roomEvent SubscribeAsync error", zap.Error(err))
	}
}

// 发布事件
func (r *roomEvent) Publish(event string, args ...any) {
	r.bus.Publish(event, args...)
}
