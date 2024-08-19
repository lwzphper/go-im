package repo

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go-im/pkg/logger"
	pkgRedis "go-im/pkg/redis"
	"go-im/pkg/util"
	"go.uber.org/zap"
)

var RoomCache = &roomCache{}

type roomCache struct{}

// IsCreate 判断房间是否创建
func (r *roomCache) IsCreate(roomId uint64) bool {
	return r.c().HExists(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// GetName 获取房间名称
func (r *roomCache) GetName(roomId uint64) string {
	return r.c().HGet(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// 获取房间列表
func (r *roomCache) List() map[string]string {
	return r.c().HGetAll(context.Background(), cacheKeyCreateRoomId).Val()
}

// 创建房间
func (r *roomCache) Create(roomId uint64, roomName string) (bool, error) {
	script := `
	local value = redis.call("HEXISTS", KEYS[1], ARGV[1])
		if( value == 0 ) then
			return redis.call("HSET" , KEYS[1],ARGV[1],ARGV[2])
		end
		return 0
`
	result, err := redis.NewScript(script).Run(context.Background(), r.c().Conn(), []string{cacheKeyCreateRoomId}, roomId, roomName).Result()
	if err != nil {
		logger.Error("create room lua script error", zap.Error(err))
		return false, errors.New("create room error")
	}
	ret := result.(int64)
	if ret == 1 {
		return true, nil
	}
	return false, nil
}

// 删除房间
func (r *roomCache) Remove(roomId uint64) int64 {
	return r.c().HDel(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// 获取 redis 客户端
func (r *roomCache) c() *redis.Client {
	return pkgRedis.C(pkgRedis.NAME_DEFAULT)
}
