package service

import (
	"fmt"
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
)

type MsgHandler func(n *connect.Node, data *types.Input)

type MsgStrategy map[types.MsgMethod]MsgHandler

// 注册
func (m MsgStrategy) Register(method types.MsgMethod, handler MsgHandler) {
	if _, ok := m[method]; ok {
		panic(fmt.Sprintf("method %d 重复注册", method.Uint8())) // 暂不支持一个方法绑定多个处理器
	}
	m[method] = handler
}

// 获取
func (m MsgStrategy) Get(method types.MsgMethod) MsgHandler {
	if h, ok := m[method]; ok {
		return h
	}

	return nil
}
