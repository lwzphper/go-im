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

func NewRoomCache() *RoomCache {
	return &RoomCache{rdClient: pkgRedis.C(pkgRedis.NAME_DEFAULT)}
}

type RoomCache struct {
	rdClient *redis.Client
}

// IsCreate 判断房间是否创建
func (r *RoomCache) IsCreate(roomId uint64) bool {
	return r.rdClient.HExists(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// GetName 获取房间名称
func (r *RoomCache) GetName(roomId uint64) string {
	return r.rdClient.HGet(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// 获取房间列表
func (r *RoomCache) List() map[string]string {
	return r.rdClient.HGetAll(context.Background(), cacheKeyCreateRoomId).Val()
}

// 创建房间。返回值：是否创建成功，0 表示房间已存在或创建失败
func (r *RoomCache) Create(roomId uint64, roomName string) (bool, error) {
	script := `
	local value = redis.call("HEXISTS", KEYS[1], ARGV[1])
		if( value == 0 ) then
			return redis.call("HSET" , KEYS[1],ARGV[1],ARGV[2])
		end
		return 0
`
	result, err := redis.NewScript(script).Run(context.Background(), r.rdClient.Conn(), []string{cacheKeyCreateRoomId}, roomId, roomName).Result()
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
func (r *RoomCache) Remove(roomId uint64) int64 {
	return r.rdClient.HDel(context.Background(), cacheKeyCreateRoomId, util.Uint64ToString(roomId)).Val()
}

// 获取 redis 客户端
/*func (r *roomCache) c() *redis.Client {
	return pkgRedis.C(pkgRedis.NAME_DEFAULT)
}*/
