package service

import (
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

type MsgHandler func(n *connect.Node, data *types.Input)

type MsgStrategy map[types.MsgMethod]MsgHandler

// 注册
func (m MsgStrategy) Register(method types.MsgMethod, handler MsgHandler) {
	/*if _, ok := m[method]; ok {
		return errors.New("msg method already exists")
	}*/
	m[method] = handler
}

// 获取
func (m MsgStrategy) Get(method types.MsgMethod) MsgHandler {
	if h, ok := m[method]; ok {
		return h
	}

	return nil
}
