package service

import (
	types2 "go-im/internal/connect"
	"go-im/internal/logic/room/types"
	"go-im/pkg/logger"
)

// 分发消息
func (s *Service) Dispatch(n *types2.Node, message []byte) {
	data, err := types.UnMarshalInput(message)
	if err != nil {
		logger.Infof("用户：%d 消息格式有误：%s", n.UserId, string(message))
		s.sendErrorMsg(n, "", types.MethodServiceNotice, types.CodeValidateError, "消息格式有误")
		return
	}

	// 设置房间id默认值
	if data.RoomId == 0 {
		data.RoomId = n.RoomId
	}

	method := s.strategy.Get(types.MsgMethod(data.Method))
	if method == nil {
		s.sendErrorMsg(n, data.RequestId, types.MethodServiceNotice, types.CodeValidateError, "method 有误")
		return
	}

	method(n, data)
}
