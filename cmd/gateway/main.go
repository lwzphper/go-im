package main

import (
	"context"
	"flag"
	consulApi "github.com/hashicorp/consul/api"
	"go-im/config"
	"go-im/internal/gateway/api"
	"go-im/internal/gateway/api/middleware"
	"go-im/internal/gateway/domain/proxy"
	_ "go-im/pkg/config"
	"go-im/pkg/logger"
	"go-im/pkg/response"
	"go-im/pkg/server"
	"go-im/pkg/util"
	"go-im/pkg/util/consul"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	var confFile = flag.String("f", "", "the service config from file")

	flag.Parse()
	defer logger.Sync()

	// 初始化配置文件
	if *confFile == "" {
		config.NewDefault().Builder()
	} else {
		config.NewFileBuilder().Builder(*confFile)
	}

	util.Validator(util.ZhLocale)

	// 初始化 consul，并监听
	consul.Init(config.C.Consul.Host)
	// 初始化代理服务（依赖 consul 服务， consul 需先初始化）
	proxy.Init()

	// 监听服务变更
	go consul.C().Watch(func(srv []*consulApi.AgentService) {
		proxy.C().ServiceChange(srv)
	})
	//go consul.C().Heartbeat()

	// 启动gin
	service := GinServer{}
	service.Start()
}

type GinServer struct{}

// Start 启动
func (s *GinServer) Start() {
	gin.SetMode(config.C.App.Env)
	r := s.setRouter()

	srv := &http.Server{
		ReadHeaderTimeout: 60 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1M
		Addr:              config.C.App.GatewayAddr,
		Handler:           r,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server listen error: %s\n", err)
		}
	}()

	hook := server.NewHook()
	hook.Close(func(sg os.Signal) {
		logger.Info("Shutdown Server ...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("Server Shutdown error:%v", err)
		}
		logger.Info("Server exiting")
	})
}

// 设置路由
func (s *GinServer) setRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/favicon.ico", func(g *gin.Context) {})

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		response.NotFoundError(c.Writer)
	})

	r.Use(middleware.Cors())      // 配置跨域
	r.Use(middleware.Exception()) // 错误处理

	// 静态页面
	r.LoadHTMLFiles("static/index.html")
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"gateway_host": config.C.GetGatewayHost(),
		})
	})

	// 代理
	r.GET("/proxy", proxy.Handle)

	g := r.Group("go-im")
	api.RegisterUser(g) // 注册用户

	return r
}
