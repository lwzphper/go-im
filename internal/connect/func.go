package connect

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go-im/pkg/logger"
	"go.uber.org/zap"
	"io"
	"strings"
	"time"
)

// WsErrorNeedClose 判断 ws 错误是否需要关闭
func WsErrorNeedClose(err error) bool {
	var closeError *websocket.CloseError
	if errors.As(err, &closeError) {
		logger.Debug("连接关闭")
		return true
	}

	str := err.Error()

	// 服务器主动关闭连接
	if strings.HasSuffix(str, "use of closed network connection") {
		return true
	}
	// 远程主机强制关闭连接
	if strings.HasSuffix(str, "An existing connection was forcibly closed by the remote host") {
		return true
	}

	if err == io.EOF {
		return true
	}

	// SetReadDeadline 之后，超时返回的错误
	if strings.HasSuffix(str, "i/o timeout") {
		return true
	}

	logger.Debug("read tcp error：", zap.Error(err))
	return false
}

// 发送成功消息
func writeTextMessage(conn *websocket.Conn, method MsgMethod, data string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if err := conn.WriteMessage(websocket.TextMessage, MarshalSystemOutput(method, data)); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}

// 发送错误消息
func OutputError(conn *websocket.Conn, code Code, msg string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if msg == "" {
		msg = code.Name()
	}

	data := Output{
		Code:   code,
		Method: MethodServiceNotice,
		Msg:    msg,
	}

	if err := conn.WriteMessage(websocket.TextMessage, data.Marshal()); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}
