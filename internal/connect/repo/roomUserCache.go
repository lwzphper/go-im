package repo

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-im/pkg/logger"
	pkgRedis "go-im/pkg/redis"
	util "go-im/pkg/util"
)

var RoomUserCache = &roomUserCache{}

type roomUserCache struct{}

// 是否存在
func (r *roomUserCache) Exists(roomId, userId uint64) bool {
	exist, err := r.c().HExists(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Result()
	if err != nil {
		logger.Errorf("CheckUserIsLogin redis error: %v", err)
		return false
	}
	return exist
}

// 创建
func (r *roomUserCache) Create(roomId, userId uint64, username string) int64 {
	return r.c().HSet(context.Background(), r.cKey(roomId), util.Uint64ToString(userId), username).Val()
}

// 获取全部数据
func (r *roomUserCache) GetAll(roomId uint64) map[string]string {
	return r.c().HGetAll(context.Background(), r.cKey(roomId)).Val()
}

// 删除
func (r *roomUserCache) Remove(roomId, userId uint64) int64 {
	return r.c().HDel(context.Background(), r.cKey(roomId), util.Uint64ToString(userId)).Val()
}

// 删除整个房间数据
func (r *roomUserCache) DeleteRoom(roomId uint64) int64 {
	return r.c().Del(context.Background(), r.cKey(roomId)).Val()
}

func (r *roomUserCache) cKey(roomId uint64) string {
	return fmt.Sprintf("room:%d", roomId)
}

// 获取 redis 客户端
func (r *roomUserCache) c() *redis.Client {
	return pkgRedis.C(pkgRedis.NAME_DEFAULT)
}
