package repo

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go-im/pkg/logger"
	pkgRedis "go-im/pkg/redis"
	"go-im/pkg/util"
	"go.uber.org/zap"
)

/**
 * @Description: 用户连接 IM service 服务id
 */

func NewUserServiceCache() *UserServiceCache {
	return &UserServiceCache{
		rdClient: pkgRedis.C(pkgRedis.NAME_DEFAULT),
	}
}

type UserServiceCache struct {
	rdClient *redis.Client
}

// 获取用户的 serviceId
func (r *UserServiceCache) get(roomId uint64, userId uint64) string {
	return r.rdClient.HGet(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Val()
}

// 创建用户与 serviceId 的映射
func (r *UserServiceCache) Create(roomId, userId uint64, serviceId string) bool {
	script := `
	local value = redis.call("HEXISTS", KEYS[1], ARGV[1])
		if( value == 0 ) then
			return redis.call("HSET" , KEYS[1], ARGV[1], ARGV[2])
		end
		return 0
`
	result, err := redis.NewScript(script).Run(context.Background(), r.rdClient.Conn(), []string{r.cKey(roomId)}, userId, serviceId).Result()
	if err != nil {
		logger.Error("create user room cache lua script error", zap.Error(err))
	}
	ret := result.(int64)
	if ret == 1 {
		return true
	}
	return false
}

// 删除用户于房间的映射
func (r *UserServiceCache) Remove(roomId, userId uint64) int64 {
	return r.rdClient.HDel(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Val()
}

// 删除整个房间的缓存
func (r *UserServiceCache) DeleteRoom(roomId uint64) int64 {
	return r.rdClient.Del(context.Background(), r.cKey(roomId)).Val()
}

// 缓存key
func (r *UserServiceCache) cKey(roomId uint64) string {
	return cacheKeyUserService + util.Uint64ToString(roomId)
}
