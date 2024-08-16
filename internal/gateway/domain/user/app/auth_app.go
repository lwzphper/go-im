package app

import (
	"github.com/gin-gonic/gin"
	"go-im/config"
	"go-im/internal/gateway/domain/user"
	"go-im/internal/gateway/domain/user/service"
	"go-im/pkg/jwt"
	"go-im/pkg/response"
	util "go-im/pkg/util"
	"time"
)

func NewAuthApp() *AuthApp {
	return &AuthApp{
		userServer: service.NewUserService(),
	}
}

type AuthApp struct {
	userServer service.IUserService
}

// LoginRegister 登录，如果账号不存在，自动注册
func (a *AuthApp) LoginRegister(c *gin.Context) {
	var req user.LoginReq
	if err := c.ShouldBind(&req); err != nil {
		util.HandleValidatorError(c, err)
		return
	}

	loginInfo, err := a.userServer.LoginRegister(c, &req)
	if err != nil {
		response.Error(c.Writer, err)
		return
	}

	a.sendLoginResult(c, loginInfo)
}

// Login 登录
func (a *AuthApp) Login(c *gin.Context) {
	var req user.LoginReq
	if err := c.ShouldBind(&req); err != nil {
		util.HandleValidatorError(c, err)
		return
	}

	loginInfo, err := a.userServer.Login(c, &req)
	if err != nil {
		response.Error(c.Writer, err)
		return
	}

	a.sendLoginResult(c, loginInfo)
}

// 发送登录结果
func (a *AuthApp) sendLoginResult(c *gin.Context, loginInfo *user.LoginResult) {
	// 生成 token
	gen := jwt.NewJwtTokenGen(config.C.App.Name, []byte(config.C.Jwt.Secret))
	token, err := gen.GenerateToken(loginInfo.Id, time.Duration(config.C.Jwt.TTL)*time.Second)
	if err != nil {
		response.Error(c.Writer, err)
		return
	}

	loginInfo.Token = token

	response.Success(c.Writer, loginInfo)
}

// Register 注册
func (a *AuthApp) Register(c *gin.Context) {
	var req user.RegisterReq
	if err := c.ShouldBind(&req); err != nil {
		util.HandleValidatorError(c, err)
		return
	}

	userId, err := a.userServer.Register(c, &req)
	if err != nil {
		response.Error(c.Writer, err)
		return
	}

	response.Success(c.Writer, user.RegisterResult{UserId: userId})
}

// GetImServer 获取 IM 服务器
func (a *AuthApp) GetImServer(c *gin.Context) {
	response.Success(c.Writer, a.userServer.GetImServer(c))
}
