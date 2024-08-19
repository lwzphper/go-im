package util

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IpHost struct {
	Host string
	Port int
}

func (h *IpHost) String() string {
	if h == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// SplitAddress 切分ip地址，获取 IP 和 端口号
func SplitAddress(addr string, inDocker bool) (*IpHost, error) {
	var (
		err  error
		port int
	)

	if addr == "" {
		return nil, fmt.Errorf("address is empty")
	}

	split := strings.Split(addr, ":")
	if len(split) != 1 && len(split) != 2 {
		return nil, fmt.Errorf("address format error: %s", addr)
	}

	port = 80
	if len(split) == 2 {
		if port, err = strconv.Atoi(split[1]); err != nil {
			return nil, fmt.Errorf("address port error %s, port: %s", err.Error(), split[1])
		}
	}

	ip := split[0]
	if ip == "" {
		if inDocker {
			ip = "host.docker.internal" // docker 访问宿主机地址
		} else {
			ip = "127.0.0.1"
		}
	}

	return &IpHost{
		Host: ip,
		Port: port,
	}, nil
}

// 判断是否ip地址
func CheckIsIp(ip string) bool {
	return net.ParseIP(ip) != nil
}

func Uint64ToString(num uint64) string {
	return strconv.FormatUint(num, 10)
}

func StringToUint64(str string) (uint64, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.ParseUint(str, 10, 64)
}
