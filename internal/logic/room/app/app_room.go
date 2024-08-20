package app

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
	"go-im/internal/logic/room/service"
	"go-im/pkg/logger"
)

func NewRoomApp() *RoomApp {
	return &RoomApp{
		roomService: service.NewService(),
		strategy:    msgStrategy{},
	}
}

type RoomApp struct {
	roomService service.IService
	strategy    msgStrategy
}

// 分发消息
func (r *RoomApp) Dispatch(n *connect.Node, message []byte) {
	data, err := room.UnMarshalInput(message)
	if err != nil {
		logger.Infof("用户：%d 消息格式有误：%s", n.UserId, string(message))
		n.DataQueue <- &room.QueueMsgData{
			Method: room.MethodServiceNotice,
			Code:   room.CodeValidateError,
			Msg:    "消息格式有误",
		}
		return
	}

	// 设置房间id默认值
	if data.RoomId == 0 {
		data.RoomId = n.RoomId
	}

	method := r.strategy.get(room.MsgMethod(data.Method))
	if method == nil {
		n.DataQueue <- &room.QueueMsgData{
			Method: room.MethodServiceNotice,
			Code:   room.CodeValidateError,
			Msg:    "method 有误",
		}
		return
	}

	method(n, data)
}

// 关闭
func (r *RoomApp) Close(n *connect.Node) {
	r.roomService.Close(n)
}

// 初始化
func (r *RoomApp) Init() *RoomApp {
	r.strategy.register(room.MethodCreateRoom, r.roomService.Create)
	r.strategy.register(room.MethodJoinRoom, r.roomService.Join)
	r.strategy.register(room.MethodRoomUser, r.roomService.UserList)
	r.strategy.register(room.MethodNormal, r.roomService.Normal)
	r.strategy.register(room.MethodGroup, r.roomService.GroupMsg)
	r.strategy.register(room.MethodRoomList, r.roomService.RoomList)
	r.strategy.register(room.MethodCreateRoomNotice, r.roomService.CreateRoomNotice)
	r.strategy.register(room.MethodOffline, r.roomService.LeaveRoom)

	return r
}
