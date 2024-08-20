package app

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room"
)

type MsgHandler func(n *connect.Node, data *room.Input)

type msgStrategy map[room.MsgMethod]MsgHandler

// 注册
func (m msgStrategy) register(method room.MsgMethod, handler MsgHandler) {
	/*if _, ok := m[method]; ok {
		return errors.New("msg method already exists")
	}*/
	m[method] = handler
}

// 获取
func (m msgStrategy) get(method room.MsgMethod) MsgHandler {
	if h, ok := m[method]; ok {
		return h
	}

	return nil
}
