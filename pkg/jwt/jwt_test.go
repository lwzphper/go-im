package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jwtSignKey = []byte("aomKPd2HzKTNuEiX5tc5IlriiUIb9IEBvAI0jMGzOQEp38yfL0cJibvQMYTLpxoX")
	wantStr    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJtaW5pLWdvIiwiZXhwIjoxNjkwOTcwMzU2LCJpYXQiOjE2OTA5NzAzNTYsImF1ZCI6MX0.-KV9Us5vwR5bE9EjiDXSnnic6bgcxn5bqSw021UHIK8"
)

func TestGenerateToken(t *testing.T) {

	cases := []struct {
		name      string
		signKey   []byte
		now       time.Time
		expireSec time.Duration
		want      string
		wantEqual bool
	}{
		{
			name:      "valid_token",
			signKey:   jwtSignKey,
			now:       time.Unix(1690970356, 0), // 固定时间，防止随机值生成的token变化
			want:      wantStr,
			wantEqual: true,
		},
		{
			name:      "fail_token",
			signKey:   jwtSignKey,
			now:       time.Now(),
			want:      wantStr,
			wantEqual: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			jwtGen := NewJwtTokenGen("demo", c.signKey)
			jwtGen.nowFunc = func() time.Time {
				return c.now
			}
			token, err := jwtGen.GenerateToken(1, c.expireSec)
			if err != nil {
				t.Errorf("token generate error: %v", err)
			}

			if c.wantEqual {
				assert.Equal(t, c.want, token)
			}

			if !c.wantEqual {
				assert.NotEqual(t, c.want, token)
			}
		})
	}
}

func TestTokenVerify(t *testing.T) {
	cases := []struct {
		name      string
		id        uint64
		expireSec time.Duration
		want      uint64
		wantEqual bool
		wantErr   bool
	}{
		{
			name:      "subject_equal",
			id:        123,
			expireSec: 10 * time.Second,
			want:      123,
			wantEqual: true,
		},
		{
			name:      "token_expire",
			id:        123,
			want:      0, // jwt 解析错误，获取不到响应数据
			wantErr:   true,
			wantEqual: true,
		},
		{
			name:      "subject_not_equal",
			id:        666,
			expireSec: 10 * time.Second,
			want:      123,
			wantEqual: false,
		},
	}

	gen := NewJwtTokenGen("demo", jwtSignKey)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			jwtToken, err := gen.GenerateToken(c.id, c.expireSec)
			if err != nil {
				t.Errorf("generate token error: %v", err)
				return
			}

			validator := NewTokenValidator(jwtSignKey)
			claims, err := validator.Validator(jwtToken)
			if !c.wantErr && err != nil {
				t.Errorf("validate token error：%v", err)
				return
			}

			var id uint64
			if claims == nil {
				id = 0
			} else {
				id = claims.Audience
			}

			if c.wantEqual {
				assert.Equal(t, c.want, id)
			} else {
				assert.NotEqual(t, c.want, id)
			}
		})
	}

}
