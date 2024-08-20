package connect

import (
	"github.com/gorilla/websocket"
	"go-im/internal/logic/room"
	"go-im/pkg/logger"
	"go.uber.org/zap"
	"time"
)

// 发送成功消息
func writeTextMessage(conn *websocket.Conn, method room.MsgMethod, data string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if err := conn.WriteMessage(websocket.TextMessage, room.MarshalSystemOutput(method, data)); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}

// 发送错误消息
func OutputError(conn *websocket.Conn, code room.Code, msg string) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logger.Error("set write deadline error", zap.Error(err))
		return false
	}

	if msg == "" {
		msg = code.Name()
	}

	data := room.Output{
		Code:   code,
		Method: room.MethodServiceNotice,
		Msg:    msg,
	}

	if err := conn.WriteMessage(websocket.TextMessage, data.Marshal()); err != nil {
		logger.Error("text msg send error", zap.Error(err))
		return false
	}
	return true
}
