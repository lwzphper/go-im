package types

import (
	"encoding/json"
	"go-im/pkg/logger"
	"go.uber.org/zap"
)

type Code int32

const (
	CodeSuccess       Code = 0
	CodeError         Code = 501
	CodeAuthError     Code = 40001
	CodeValidateError Code = 40002
)

var CodeName = map[Code]string{
	CodeSuccess:       "ok",
	CodeError:         "系统繁忙，请稍后再试。",
	CodeAuthError:     "授权失败，请重新登录",
	CodeValidateError: "数据验证错误",
}

func (c Code) Name() string {
	if val, ok := CodeName[c]; ok {
		return val
	}
	return "未知错误"
}

type MsgMethod uint8

func (m MsgMethod) Uint8() uint8 {
	return uint8(m)
}

// input method
const (
	MethodCreateRoom       MsgMethod = iota + 1 // 创建房间
	MethodJoinRoom                              // 加入房间
	MethodRoomList                              // 房间列表
	MethodRoomUser                              // 获取房间用户列表
	MethodGroup                                 // 群聊消息
	MethodNormal                                // 普通消息
	MethodOnline                                // 上线消息/加入房间
	MethodOffline                               // 下线消息
	MethodCreateRoomNotice                      // 新增房间通知
)

// Service method
const (
	MethodServiceNotice MsgMethod = iota + 100 // 服务器的消息（如：消息错误、通知等）
	MethodServiceAck                           // 确认消息
	MethodNewRoomNotice                        // 新建房间通知
)

// 队列数据
type QueueMsgData struct {
	RequestId    string    `json:"request_id,omitempty"`
	Method       MsgMethod `json:"method"`
	Code         Code      `json:"code"`
	Msg          string    `json:"msg"`
	FromUid      uint64    `json:"from_uid,omitempty"`
	FromUsername string    `json:"from_username,omitempty"` // 消息发送者名称
	Data         any       `json:"data"`
	RoomId       uint64    `json:"room_id,omitempty"`     // 房间id
	ToUid        uint64    `json:"to_uid,omitempty"`      // 消息接收者
	FromServer   string    `json:"from_server,omitempty"` // 消息来源（可能为空，广播时使用）
}

func (q *QueueMsgData) Marshal() []byte {
	result, _ := json.Marshal(q)
	return result
}

// 转换输出结果
func (q *QueueMsgData) MarshalOutput(roomId uint64) []byte {
	data := Output{
		RequestId:    q.RequestId,
		Code:         q.Code,
		Msg:          q.Msg,
		Method:       q.Method,
		Data:         q.Data,
		FromUid:      q.FromUid,
		FromUsername: q.FromUsername,
		RoomId:       roomId,
		FromServer:   q.FromServer,
	}

	if data.Msg == "" {
		data.Msg = data.Code.Name()
	}
	return data.Marshal()
}

// 下行数据
type Output struct {
	RequestId    string    `json:"request_id"`              // 消息id
	Code         Code      `json:"code"`                    // 状态码
	Msg          string    `json:"msg,omitempty"`           // 错误信息
	Method       MsgMethod `json:"method"`                  // 调用方法
	Data         any       `json:"data"`                    // 传递的消息
	FromUid      uint64    `json:"from_uid,omitempty"`      // 消息发送者
	FromUsername string    `json:"from_username,omitempty"` // 消息发送者名称
	RoomId       uint64    `json:"room_id,omitempty"`       // 房间id
	ToUid        uint64    `json:"to_uid,omitempty"`        // 消息接收者
	FromServer   string    `json:"from_server,omitempty"`   // 消息来源（广播时使用）
}

func (w *Output) QueueMsgData() *QueueMsgData {
	data := QueueMsgData{
		Code:         w.Code,
		Msg:          w.Msg,
		RequestId:    w.RequestId,
		Method:       w.Method,
		FromUid:      w.FromUid,
		FromUsername: w.FromUsername,
		Data:         w.Data,
		RoomId:       w.RoomId,
		FromServer:   w.FromServer,
	}
	return &data
}

func (w *Output) Marshal() []byte {
	result, _ := json.Marshal(w)
	return result
}

// MarshalOutput 序列化消息
func MarshalOutput(m MsgMethod, data string, FromUid uint64) []byte {
	result := Output{
		Method:  m,
		Data:    data,
		FromUid: FromUid,
	}
	return result.Marshal()
}

// 上行数据
type Input struct {
	RequestId string `json:"request_id,omitempty"` // 消息id
	Method    uint8  `json:"method"`               // 调用方法
	Data      any    `json:"data"`                 // 传递的消息
	RoomId    uint64 `json:"room_id,omitempty"`    // 房间id
	ToUid     uint64 `json:"to_uid,omitempty"`     // 消息接收者
	FromUid   uint64 `json:"from_uid,omitempty"`   // 消息代理时传递
}

func (i *Input) Marshal() []byte {
	result, _ := json.Marshal(i)
	return result
}

// MarshalSystemOutput 序列化消息系统下行消息
func MarshalSystemOutput(m MsgMethod, msg string) []byte {
	return MarshalOutput(m, msg, 0)
}

// UnMarshalInput 解析上行消息
func UnMarshalInput(data []byte) (*Input, error) {
	var ret = new(Input)
	if err := json.Unmarshal(data, ret); err != nil {
		logger.Error("input data marshal error", zap.Error(err))
		return nil, err
	}

	return ret, nil
}

// 用户列表
type UserList []UserItem

func (u *UserList) Marshal() string {
	result, err := json.Marshal(u)
	if err != nil {
		logger.Error("userList marshal error", zap.Error(err))
		return ""
	}
	return string(result)
}

type UserItem struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

// 创建房间结果
type CreateRoomResult struct {
	RoomId uint64 `json:"room_id"`
}

func (c *CreateRoomResult) Marshal() string {
	result, err := json.Marshal(c)
	if err != nil {
		logger.Error("CreateRoomResult marshal error", zap.Error(err))
		return ""
	}
	return string(result)
}

type RoomList []RoomInfo

type RoomInfo struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func (i *RoomList) Marshal() string {
	result, err := json.Marshal(i)
	if err != nil {
		logger.Error("roomList marshal error", zap.Error(err))
		return ""
	}
	return string(result)
}
