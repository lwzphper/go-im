package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go-im/internal/connect"
	types2 "go-im/internal/logic/room/types"
	"go-im/pkg/logger"
)

// 网管消息
func (s *Service) GatewayMsg(wsConn *websocket.Conn, message []byte) {
	var data = new(types2.QueueMsgData)
	err := json.Unmarshal(message, data)
	if err != nil {
		logger.Infof("网关消息格式有误：%s", string(message))
		connect.WriteTextMessage(wsConn, types2.MethodServiceNotice, "消息格式有误")
		return
	}

	switch data.Method {
	case types2.MethodNormal: // 普通消息。发送指定用户
		if node := connect.GetNode(data.ToUid); node != nil {
			node.DataQueue <- data.MarshalOutput(node.RoomId)
		}
	case types2.MethodCreateRoomNotice: // 创建房间
		connect.PushAll(data)
	default:
		// 推送当前服务指定房间的全部用户
		s.SendRoomMsg(data.RoomId, data)
	}
}
