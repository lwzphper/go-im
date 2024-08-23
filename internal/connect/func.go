package connect

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go-im/internal/event"
	"go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

const (
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second // 写超时时间
	pingPeriod = 60 * time.Second // 心跳时间周期

)

// 发送文本消息
func WriteTextMessage(conn *websocket.Conn, method types.MsgMethod, data string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if err := conn.WriteMessage(websocket.TextMessage, types.MarshalSystemOutput(method, data)); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}

// 发送错误消息
func OutputError(conn *websocket.Conn, code types.Code, msg string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if msg == "" {
		msg = code.Name()
	}

	data := types.Output{
		Code:   code,
		Method: types.MethodServiceNotice,
		Msg:    msg,
	}

	if err := conn.WriteMessage(websocket.TextMessage, data.Marshal()); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}

// 读取conn错误。返回结果：是否需要终止监听
func handleReadErr(n *Node, err error) bool {
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) {
		logger.Debug("连接关闭", zap.String("用户id", util.Uint64ToString(n.UserId)))
		CloseConn(n)
		return true
	}

	str := err.Error()

	// 服务器主动关闭连接
	if strings.HasSuffix(str, "use of closed network connection") {
		CloseConn(n)
		return true
	}

	if err == io.EOF {
		CloseConn(n)
		return true
	}

	// SetReadDeadline 之后，超时返回的错误
	if strings.HasSuffix(str, "i/o timeout") {
		CloseConn(n)
		return true
	}

	logger.Debug("read tcp error：", zap.Uint64("user_id", n.UserId), zap.Error(err))
	return false
}

// 关闭连接
func CloseConn(n *Node) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()

	if n.IsClose {
		return
	}
	n.IsClose = true

	if err := n.Conn.Close(); err != nil {
		logger.Error("close node connect error", zap.Error(err))
	}
	logger.Debugf("关闭用户连接成功：%d", n.UserId)
	close(n.DataQueue)
	close(n.BroadcastQueue)

	// 删除用户连接映射
	DeleteNode(n.UserId)

	event.RoomEvent.Publish(event.CloseConn, n)

	logger.Debugf("用户：%d 程序已关闭连接", n.UserId)
}
