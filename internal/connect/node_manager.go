package connect

import (
	"go-im/internal/logic/room/types"
	types2 "go-im/internal/types"
	"sync"
)

// 全部用户（不区分群组）
var NodesManger = sync.Map{} // UserId => *Node

// GetNode 获取用户连接
func GetNode(userId uint64) *types2.Node {
	if value, ok := NodesManger.Load(userId); ok {
		return value.(*types2.Node)
	}
	return nil
}

// SetNode 设置用户连接
func SetNode(userId uint64, conn *types2.Node) {
	NodesManger.Store(userId, conn)
}

// DeleteNode 删除用户连接
func DeleteNode(userId uint64) {
	NodesManger.Delete(userId)
}

// 广播消息
func PushAll(data *types.QueueMsgData) {
	NodesManger.Range(func(key, value any) bool {
		node := value.(*types2.Node)
		node.DataQueue <- data
		return true
	})
}
