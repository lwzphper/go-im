package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConsul_watch(t *testing.T) {
	err := RegisterWatcher(WatchTypeServices, nil, "127.0.0.1:8500", func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case []*api.ServiceEntry:
			for _, i := range d {
				// 这里是单个service变化时需要做的逻辑，可以自己添加，或在外部写一个类似handler的函数传进来
				fmt.Printf("service %s 已变化", i.Service.Service)
				// 打印service的状态
				fmt.Println("service status: ", i.Checks.AggregatedStatus())
			}
		}
	})
	assert.Nil(t, err)
}
