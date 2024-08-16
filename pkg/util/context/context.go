package context

import (
	"github.com/gin-gonic/gin"
)

var (
	ctxUserIdKey = "ctxUserId" // 登录人用户id
)

// CtxWithUserID 设置用户id
func CtxWithUserID(c *gin.Context, id uint64) {
	c.Set(ctxUserIdKey, id)
}

// UserIDFromCtx 获取用户id
func UserIDFromCtx(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(ctxUserIdKey)
	if !ok {
		return 0, false
	}

	return value.(uint64), true
}
