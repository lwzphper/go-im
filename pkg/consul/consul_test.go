package consul

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

// 测试http服务（先开启再执行 TestConsul_Register）
func TestHttpService(t *testing.T) {
	startHttp()
}

func TestConsul_Register(t *testing.T) {
	c := NewClient("127.0.0.1:8500")
	regName := "test-im1"
	host := "host.docker.internal" // docker 访问宿主机地址

	err := c.Register(host, 8080, regName, "test-server1", []string{"im-service"})
	assert.Nil(t, err)
	err = c.Register(host, 8081, regName, "test-server2", []string{"im-service"})
	assert.Nil(t, err)

	regName2 := "test-im2"

	err = c.Register(host, 8082, regName2, "test-server3", []string{"im-service"})
	assert.Nil(t, err)
	err = c.Register(host, 8083, regName2, "test-server4", []string{"im-service"})
	assert.Nil(t, err)

	healthService, err := c.HealthService(regName, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(healthService))

	// 全部服务
	lists, err := c.AllService()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(lists))

	// 指定Id获取服务
	service, err := c.Service("test-server1")
	assert.Nil(t, err)
	assert.Equal(t, "test-server1", service.ID)

	// 指定通过服务名查找服务
	services, err := c.ServiceByName(regName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(services))
}

func startHttp() {
	//定义一个http接口
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("OK")); err != nil {
			fmt.Println("error: ", err.Error())
		}
	})
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("error: ", err.Error())
	}
}
