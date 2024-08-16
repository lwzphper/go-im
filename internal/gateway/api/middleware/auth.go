package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-im/config"
	"go-im/pkg/errorx"
	"go-im/pkg/jwt"
	"go-im/pkg/response"
	"go-im/pkg/util/context"
)

var (
	ErrTokenUnauthorized = errorx.New(response.CodeUnauthorized, "token不合法或不存在", "授权失败，请求重新登录！")
	ErrTokenExpire       = errorx.New(response.CodeUnauthorized, "token登录过期", "登陆已过期，请重新登陆！")
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handleParseToken(c); err != nil {
			response.Dynamic(c.Writer, nil, err)
			c.Abort()
		}

		c.Next()
	}
}

func handleParseToken(c *gin.Context) error {
	token := c.GetHeader("Authorization")
	if token == "" {
		return ErrTokenUnauthorized
	}

	validator := jwt.NewTokenValidator([]byte(config.C.Jwt.Secret))
	claims, err := validator.Validator(token)

	if err != nil || claims == nil {
		if errors.Is(err, jwt.ErrExpiredOrNotValid) {
			return ErrTokenExpire
		} else {
			return ErrTokenUnauthorized
		}
	}

	// 防止用户id为0的情况
	if claims.Audience == 0 {
		return ErrTokenUnauthorized
	}

	// 设置 用户id
	context.CtxWithUserID(c, claims.Audience)

	return nil
}
