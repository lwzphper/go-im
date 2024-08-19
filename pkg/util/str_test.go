package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitAddress(t *testing.T) {
	cases := []struct {
		name     string
		addr     string
		wantHost string
		wantPort int
		inDocker bool
		wantErr  bool
	}{
		{
			name:     "address with port",
			addr:     "127.0.0.1:8080",
			wantHost: "127.0.0.1",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "address with port",
			addr:     "mysql:8080",
			wantHost: "mysql",
			wantPort: 8080,
			wantErr:  false,
		},
		{
			name:     "address without port",
			addr:     "127.0.0.1",
			wantHost: "127.0.0.1",
			wantPort: 80,
			wantErr:  false,
		},
		{
			name:    "address with invalid port",
			addr:    "127.0.0.1:8080a",
			wantErr: true,
		},
		{
			name:     "in docker",
			addr:     ":8080",
			wantPort: 8080,
			wantHost: "host.docker.internal",
			inDocker: true,
			wantErr:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ipHost, err := SplitAddress(c.addr, c.inDocker)
			if c.wantErr {
				if err == nil {
					t.Errorf("want error, got nil")
				}
				return
			}
			assert.Equal(t, c.wantHost, ipHost.Host)
			assert.Equal(t, c.wantPort, ipHost.Port)
		})
	}
}

func TestCheckIsIp(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{
			name: "ipv4",
			ip:   "127.0.0.1",
			want: true,
		},
		{
			name: "ipv6",
			ip:   "::1",
			want: true,
		},
		{
			name: "invalid ip1",
			ip:   "127.0.0.1:8080",
			want: false,
		},
		{
			name: "invalid ip2",
			ip:   "host.docker.internal",
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if CheckIsIp(c.ip) != c.want {
				t.Errorf("want error %v, got %v", c.want, !c.want)
			}
		})
	}
}
