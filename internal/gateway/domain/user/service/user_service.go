package service

import (
	"context"
	"go-im/internal/gateway/domain/user"
	"go-im/internal/gateway/domain/user/model"
	"go-im/internal/gateway/domain/user/repo"
	"go-im/pkg/util/consul"
	"golang.org/x/crypto/bcrypt"
)

type IUserService interface {
	Register(ctx context.Context, req *user.RegisterReq) (uint64, error)
	Login(ctx context.Context, req *user.LoginReq) (*user.LoginResult, error)
	LoginRegister(ctx context.Context, req *user.LoginReq) (*user.LoginResult, error)
	GetImServer(ctx context.Context) *user.ImServerResult
}

func NewUserService() IUserService {
	return &UserService{
		userRepo: repo.NewUserRepo(),
	}
}

type UserService struct {
	userRepo *repo.UserRepo
}

// Register 注册
func (u *UserService) Register(ctx context.Context, req *user.RegisterReq) (uint64, error) {
	password, err := u.hashPassword(req.Password)
	if err != nil {
		return 0, user.ErrPasswordEncrypt
	}

	// 检查账号是否存在
	exist, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return 0, err
	}
	if exist != nil {
		return 0, user.ErrUsernameExist
	}

	return u.userRepo.Add(ctx, model.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: password,
	})
}

// Login 登录
func (u *UserService) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResult, error) {
	userInfo, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if userInfo == nil {
		return nil, user.ErrUsernameNotFound
	}

	// 校验密码
	if !u.comparePasswords(userInfo.Password, req.Password) {
		return nil, user.ErrPassword
	}

	return &user.LoginResult{
		Id:            userInfo.Id,
		Username:      userInfo.Username,
		Nickname:      userInfo.Nickname,
		ServerAddress: consul.C().RoundHealthServerUrl(),
	}, nil
}

// LoginRegister 登录注册
func (u *UserService) LoginRegister(ctx context.Context, req *user.LoginReq) (*user.LoginResult, error) {
	userInfo, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// 用户不存在，直接注册
	if userInfo == nil {
		uid, err := u.Register(ctx, &user.RegisterReq{
			Username: req.Username,
			Nickname: "",
			Password: req.Password,
		})
		if err != nil {
			return nil, err
		}
		return &user.LoginResult{
			Id:            uid,
			Username:      req.Username,
			Nickname:      "",
			ServerAddress: consul.C().RoundHealthServerUrl(),
		}, nil
	}

	// 校验密码
	if !u.comparePasswords(userInfo.Password, req.Password) {
		return nil, user.ErrPassword
	}

	return &user.LoginResult{
		Id:            userInfo.Id,
		Username:      userInfo.Username,
		Nickname:      userInfo.Nickname,
		ServerAddress: consul.C().RoundHealthServerUrl(),
	}, nil
}

// GetImServer 获取 IM 服务器
func (u *UserService) GetImServer(ctx context.Context) *user.ImServerResult {
	return &user.ImServerResult{ServerAddress: consul.C().RoundHealthServerUrl()}
}

// 创建密码
func (u *UserService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// 比较密码
func (u *UserService) comparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
