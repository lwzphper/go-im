# 构建 gateway 镜像
docker build -t goim/gateway -f ./deployment/gateway/Dockerfile .
# 构建 service 镜像
docker build -t goim/service -f ./deployment/service/Dockerfile .