package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:16379",
		Password:     "", // no password set
		DB:           0,  // use default DB
		MinIdleConns: 3,
		PoolSize:     32,
		MaxRetries:   5,
	})
	_ = client.Set(context.Background(), "vvv1111", "1234", 0)
	result, err := client.Get(context.Background(), "vvv1111").Result()
	assert.Nil(t, err)
	fmt.Print(result)
}
