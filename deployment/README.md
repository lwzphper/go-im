# 构建 gateway 镜像
docker build -t goim/gateway -f ./gateway/Dockerfile .
# 构建 service 镜像
docker build -t goim/service -f ./service/Dockerfile .