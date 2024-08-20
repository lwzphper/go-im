package api

import (
	"github.com/gin-gonic/gin"
	"go-im/internal/gateway/api/middleware"
	"go-im/internal/logic/user/app"
)

func RegisterUser(r *gin.RouterGroup) {
	// 授权
	authGroup := r.Group("auth")
	{
		auth := app.NewAuthApp()
		authGroup.POST("/register", auth.Register)            // 注册
		authGroup.POST("/login", auth.Login)                  // 登录
		authGroup.POST("/login-register", auth.LoginRegister) // 登录并注册

		authGroup.Use(middleware.JwtAuth()).GET("/service", auth.GetImServer) // 获取服务器地址
	}
}
