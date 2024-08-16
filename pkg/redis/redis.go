package redis

import (
	"fmt"
	"github.com/redis/go-redis/v9"
)

var clientMap = make(map[string]*redis.Client)

const NAME_DEFAULT = "default"

type Config struct {
	Name        string
	Addr        string
	Password    string
	Database    int
	MinIdleConn int
	PoolSize    int
	MaxRetries  int
}

func NewRedis(cfg Config) *Redis {
	return &Redis{
		cfg,
	}
}

type Redis struct {
	Config
}

// C 获取客户端
func C(name string) *redis.Client {
	if name == "" {
		name = NAME_DEFAULT
	}

	if v, ok := clientMap[name]; ok {
		return v
	}

	panic(fmt.Sprintf("请先初始化redis：%s", name))
}

// Init 初始化
func (r *Redis) Init() {
	if r.Name == "" {
		r.Name = NAME_DEFAULT
	}

	client := redis.NewClient(&redis.Options{
		Addr:         r.Addr,
		Password:     r.Password, // no password set
		DB:           r.Database, // use default DB
		MaxRetries:   r.MaxRetries,
		PoolSize:     r.PoolSize,
		MinIdleConns: r.MinIdleConn,
	})
	//logger.Debugf("初始化redis：%s", r.Addr)
	//client.FlushAll(context.Background()) // 测试使用，调试完需要关闭
	clientMap[r.Name] = client
}
