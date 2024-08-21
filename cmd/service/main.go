package main

import (
	"flag"
	"go-im/config"
	"go-im/internal/connect"
	"go-im/internal/types"
	"go-im/pkg/logger"
	"go-im/pkg/server"
	"go-im/pkg/util/consul"
	"os"
)

func main() {
	var addr = flag.String("a", ":8080", "http service address")
	var confFile = flag.String("f", "", "the service config from file")

	flag.Parse()
	defer logger.Sync()

	// 初始化配置文件
	if *confFile == "" {
		config.NewDefault().Builder()
	} else {
		config.NewFileBuilder().Builder(*confFile)
	}

	// 初始化 consul
	consul.Init(config.C.Consul.Host)

	// 解决 internal 目录不能外部引用问题
	var conn *connect.WsConn
	conn = connect.InitServer(*addr)

	hook := server.NewHook()
	hook.Close(func(sg os.Signal) {
		logger.Info("Shutdown Server ...")
		conn.Close()
		// 关闭当前服务的全部连接
		connect.NodesManger.Range(func(userId, node any) bool {
			connect.CloseConn(node.(*types.Node))
			return true
		})
		logger.Info("Server exiting")
	})
}
