package types

import (
	"github.com/gorilla/websocket"
	"go-im/internal/logic/room/types"
	"sync"
	"time"
)

const (
	MsgDefaultChannelSize = 1000 // 默认消息队列大小
	HeartBeatMaxErrorNum  = 2    // 心跳允许最大错误次数
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type NodeOpt func(node *Node)

func WithNodeLoginTime(t int64) NodeOpt {
	return func(node *Node) {
		node.LoginTime = t
	}
}

type Node struct {
	Mutex           sync.Mutex               // WS互斥锁
	Conn            *websocket.Conn          // websocket连接
	UserId          uint64                   // 用户ID
	RoomId          uint64                   // 订阅的房间ID
	HeartbeatTime   int64                    // 心跳时间
	HeartbeatErrNum uint8                    // 心跳错误次数
	LoginTime       int64                    // 登录时间
	DataQueue       chan *types.QueueMsgData // 消息队列
	BroadcastQueue  chan *types.Output       // 广播消息
	ServerAddr      string                   // 服务器地址
	ServerId        string                   // 服务器ID
	IsClose         bool                     // 是否已关闭
}

func NewNode(conn *websocket.Conn, userId uint64, serverAddr, ServerId string, opts ...NodeOpt) *Node {
	nowTime := time.Now().Unix()
	node := &Node{
		Conn:           conn,
		UserId:         userId,
		HeartbeatTime:  nowTime,
		LoginTime:      nowTime,
		DataQueue:      make(chan *types.QueueMsgData, MsgDefaultChannelSize),
		BroadcastQueue: make(chan *types.Output, MsgDefaultChannelSize),
		ServerAddr:     serverAddr,
		ServerId:       ServerId,
	}

	for _, opt := range opts {
		opt(node)
	}

	return node
}

// 检查是否连接是否存活
func (n *Node) IsHeartbeatDeal() bool {
	return n.HeartbeatErrNum >= HeartBeatMaxErrorNum
}

// GetAddr 获取ip地址
func (n *Node) GetAddr() string {
	return n.Conn.RemoteAddr().String()
}
