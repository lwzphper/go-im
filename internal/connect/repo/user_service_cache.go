package repo

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go-im/pkg/logger"
	pkgRedis "go-im/pkg/redis"
	util "go-im/pkg/util"
	"go.uber.org/zap"
)

// 用户与 serviceId 的映射

var UserServiceCache = &userServiceCache{}

type userServiceCache struct{}

// 获取用户的 serviceId
func (r *userServiceCache) get(roomId uint64, userId uint64) string {
	return r.c().HGet(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Val()
}

// 创建用户与 serviceId 的映射
func (r *userServiceCache) Create(roomId, userId uint64, serviceId string) bool {
	script := `
	local value = redis.call("HEXISTS", KEYS[1], ARGV[1])
		if( value == 0 ) then
			return redis.call("HSET" , KEYS[1], ARGV[1], ARGV[2])
		end
		return 0
`
	result, err := redis.NewScript(script).Run(context.Background(), r.c().Conn(), []string{r.cKey(roomId)}, userId, serviceId).Result()
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
func (r *userServiceCache) Remove(roomId, userId uint64) int64 {
	return r.c().HDel(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Val()
}

// 删除整个房间的缓存
func (r *userServiceCache) DeleteRoom(roomId uint64) int64 {
	return r.c().Del(context.Background(), r.cKey(roomId)).Val()
}

// 缓存key
func (r *userServiceCache) cKey(roomId uint64) string {
	return cacheKeyUserService + util.Uint64ToString(roomId)
}

// 获取 redis 客户端
func (r *userServiceCache) c() *redis.Client {
	return pkgRedis.C(pkgRedis.NAME_DEFAULT)
}

// 创建房间id
/*func (r *userServiceCache) createId() uint64 {
	return uint64(pkgRedis.C(pkgRedis.NAME_DEFAULT).Incr(context.Background(), cacheKeyRoomIdIncr).Val())
}*/
