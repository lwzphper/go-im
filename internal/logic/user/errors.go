package user

import "go-im/pkg/errorx"

var (
	ErrDBOperate        = errorx.New(40001, "数据库操作异常", "服务器繁忙，请稍后再试")
	ErrPasswordEncrypt  = errorx.New(40002, "密码加密失败，可能是密码长度超过了限制", "密码校验失败")
	ErrUsernameExist    = errorx.New(40003, "账号已存在，重复添加", "账号已存在")
	ErrPassword         = errorx.New(40004, "密码校验失败", "密码有误")
	ErrUsernameNotFound = errorx.New(40005, "登录账号不存在", "账号不存在")
)
