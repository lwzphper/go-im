package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go-im/internal/connect"
	types2 "go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go.uber.org/zap"
)

// 网关消息
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
	case types2.MethodForceOfflineBroadcast: // 强制线下通知
		if mapNode := connect.GetNode(data.FromUid); mapNode != nil {
			// 由于发送方服务器在连接层已经处理，因此不需要处理，防止删除发送方服务器未登录的账号
			if data.FromServer == mapNode.ServerId {
				return
			}
			logger.Debug("强制用户下线", zap.Uint64("user_id", data.FromUid))
			connect.OutputError(mapNode.Conn, types2.CodeAuthError, "当前账号已被其他用户登录")
			connect.CloseConn(mapNode)
			return
		}
	default:
		// 推送当前服务指定房间的全部用户
		s.SendRoomMsg(data.RoomId, data)
	}
}
