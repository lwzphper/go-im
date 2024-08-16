# 启动编译环境
FROM golang:1.21.6-alpine3.19 AS builder

# 配置编译环境
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 拷贝源代码到镜像中
COPY . /go/src/go-im

# 编译
WORKDIR /go/src/go-im
RUN go mod tidy
RUN go build -o /go/bin/service ./cmd/service/main.go


FROM alpine:3.13
COPY --from=builder /go/bin/service /bin/service
ENV ADDR=:8001

# 申明暴露的端口
EXPOSE 8001

# 设置服务入口
ENTRYPOINT [ "/bin/service" ]
