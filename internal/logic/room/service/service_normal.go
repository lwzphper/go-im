package service

import (
	roomType "go-im/internal/logic/room/types"
	"go-im/internal/types"
)

// 一对一消息
func (s *Service) normal(n *types.Node, data *roomType.Input) {
	if data.ToUid == 0 {
		s.sendErrorMsg(n, data.RequestId, roomType.MethodNormal, roomType.CodeValidateError, "未选择发送的用户")
		return
	}

	s.sendErrorMsg(n, data.RequestId, roomType.MethodNormal, roomType.CodeError, "目前只支持群聊，暂不支持私聊")
	return
}
