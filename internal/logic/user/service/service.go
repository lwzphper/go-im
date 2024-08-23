package service

import (
	"context"
	"fmt"
	"go-im/internal/connect"
	"go-im/internal/logic/room/types"
	user2 "go-im/internal/logic/user"
	"go-im/internal/logic/user/model"
	"go-im/internal/logic/user/repo"
	"go-im/pkg/cache"
	"go-im/pkg/logger"
	"go-im/pkg/util/consul"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
)

var _ IService = (*Service)(nil)

type IService interface {
	Register(ctx context.Context, req *user2.RegisterReq) (uint64, error)
	Login(ctx context.Context, req *user2.LoginReq) (*user2.LoginResult, error)
	LoginRegister(ctx context.Context, req *user2.LoginReq) (*user2.LoginResult, error)
	GetImServer(ctx context.Context) *user2.ImServerResult

	UserIdName(userId uint64) string
}

func NewUserService() IService {
	return &Service{
		userRepo:      repo.NewUserRepo(),
		f:             singleflight.Group{},
		userNameCache: cache.NewLruList(1000),
	}
}

type Service struct {
	userRepo      *repo.UserRepo
	f             singleflight.Group
	userNameCache *cache.LruCache
}

// Register 注册
func (u *Service) Register(ctx context.Context, req *user2.RegisterReq) (uint64, error) {
	password, err := u.hashPassword(req.Password)
	if err != nil {
		return 0, user2.ErrPasswordEncrypt
	}

	// 检查账号是否存在
	exist, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return 0, err
	}
	if exist != nil {
		return 0, user2.ErrUsernameExist
	}

	return u.userRepo.Add(ctx, model.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: password,
	})
}

// Login 登录
func (u *Service) Login(ctx context.Context, req *user2.LoginReq) (*user2.LoginResult, error) {
	userInfo, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if userInfo == nil {
		return nil, user2.ErrUsernameNotFound
	}

	// 校验密码
	if !u.comparePasswords(userInfo.Password, req.Password) {
		return nil, user2.ErrPassword
	}

	// 通知已登录用户下线
	return u.handleLoginAfter(userInfo), nil
}

// LoginRegister 登录注册
func (u *Service) LoginRegister(ctx context.Context, req *user2.LoginReq) (*user2.LoginResult, error) {
	userInfo, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// 用户不存在，直接注册
	if userInfo == nil {
		uid, err := u.Register(ctx, &user2.RegisterReq{
			Username: req.Username,
			Nickname: "",
			Password: req.Password,
		})
		if err != nil {
			return nil, err
		}
		return &user2.LoginResult{
			Id:            uid,
			Username:      req.Username,
			Nickname:      "",
			ServerAddress: consul.C().RoundHealthServerUrl(),
		}, nil
	}

	// 校验密码
	if !u.comparePasswords(userInfo.Password, req.Password) {
		return nil, user2.ErrPassword
	}

	return u.handleLoginAfter(userInfo), nil
}

// 登录后事件
func (u *Service) handleLoginAfter(userInfo *model.User) *user2.LoginResult {
	srv := consul.C().RoundHealthServer()
	if srv != nil {
		u.forceOfflineNotify(srv.ID, userInfo.Id)
	}

	return &user2.LoginResult{
		Id:            userInfo.Id,
		Username:      userInfo.Username,
		Nickname:      userInfo.Nickname,
		ServerAddress: consul.C().FormatServerUrl(srv),
	}
}

// 通知强制下线
func (u *Service) forceOfflineNotify(serverId string, userId uint64) {
	data := types.QueueMsgData{
		Method:     types.MethodForceOfflineBroadcast,
		FromUid:    userId,
		FromServer: serverId,
	}
	connect.SendGatewayMsg(data.Marshal())
}

// GetImServer 获取 IM 服务器
func (u *Service) GetImServer(ctx context.Context) *user2.ImServerResult {
	return &user2.ImServerResult{ServerAddress: consul.C().RoundHealthServerUrl()}
}

// UserIdName 获取用户名称
func (u *Service) UserIdName(userId uint64) string {
	if name, ok := u.userNameCache.Get(userId); ok {
		return name.(string)
	}

	key := fmt.Sprintf("user_id:%d", userId)
	v, err, _ := u.f.Do(key, func() (any, error) {
		return u.userRepo.GetById(context.Background(), userId)
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
	u.userNameCache.Put(userId, username)
	return username
}

// 创建密码
func (u *Service) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// 比较密码
func (u *Service) comparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
