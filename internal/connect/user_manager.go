package connect

import (
	"context"
	"fmt"
	"go-im/internal/gateway/domain/user/model"
	"go-im/internal/gateway/domain/user/repo"
	"go-im/pkg/logger"
	"golang.org/x/sync/singleflight"
)

var UserManger = &userManger{
	f:          singleflight.Group{},
	userIdName: make(map[uint64]string),
}

type userManger struct {
	f          singleflight.Group
	userIdName map[uint64]string
	userRepo   *repo.UserRepo
}

// 获取用户名称（返回 空字符 说明用户不存在或服务报错）
func (u *userManger) name(userId uint64) string {
	if name, ok := u.userIdName[userId]; ok && name != "" {
		return name
	}

	key := fmt.Sprintf("user_id:%d", userId)
	v, err, _ := u.f.Do(key, func() (any, error) {
		return u.getUserRepo().GetById(context.Background(), userId)
	})
	if err != nil || v == nil {
		logger.Errorf("get user info error: %v", err)
		return ""
	}

	userInfo := v.(*model.User)
	username := userInfo.Username
	if userInfo.Nickname != "" {
		username = userInfo.Nickname
	}
	u.userIdName[userId] = username
	return username
}

func (u *userManger) getUserRepo() *repo.UserRepo {
	if u.userRepo == nil {
		u.userRepo = repo.NewUserRepo()
	}
	return u.userRepo
}
