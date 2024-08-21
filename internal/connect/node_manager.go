package connect

import (
	"go-im/internal/logic/room/types"
	"sync"
)

// 全部用户（不区分群组）
var NodesManger = sync.Map{} // UserId => *Node

// GetNode 获取用户连接
func GetNode(userId uint64) *Node {
	if value, ok := NodesManger.Load(userId); ok {
		return value.(*Node)
	}
	return nil
}

// SetNode 设置用户连接
func SetNode(userId uint64, conn *Node) {
	NodesManger.Store(userId, conn)
}

// DeleteNode 删除用户连接
func DeleteNode(userId uint64) {
	NodesManger.Delete(userId)
}

// 广播消息
func PushAll(data *types.QueueMsgData) {
	NodesManger.Range(func(key, value any) bool {
		node := value.(*Node)
		node.DataQueue <- data.MarshalOutput(node.RoomId)
		return true
	})
}
