#!/bin/bash

cd docker/
docker-compose up -d

cd ../
go mod tidy

rm -f dist/service
echo "开始打包 server"
go build -o ./dist/service ./cmd/service/main.go
echo "打包 server 成功"
pkill service
echo "停止 server 服务"
nohup ./dist/service -a :8080 & 2>/dev/null
echo "启动 server 8080 服务"
nohup ./dist/service -a :8081 & 2>/dev/null
echo "启动 server 8081 服务"

rm -f dist/gateway
echo "开始打包 gateway"
go build -o ./dist/gateway ./cmd/gateway/main.go
echo "打包 gateway 成功"
pkill gateway
echo "停止 gateway 服务"
nohup ./dist/gateway & 2>/dev/null
echo "启动 gateway 服务"