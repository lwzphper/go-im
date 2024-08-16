package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"

	"github.com/rs/xid"
)

const (
	// bearerWord the bearer key word for authorization
	// bearerWord string = "Bearer"

	// bearerFormat authorization token format
	// bearerFormat string = "Bearer %s"

	// reason holds the error reason.
	reason string = "UNAUTHORIZED"
)

var (
	ErrToken = errors.New("JWT token error")
	//ErrMissingJwtToken   = errorx.Unauthorized(reason, "JWT token is missing")
	ErrExpiredOrNotValid = errors.New("JWT token expire or not valid")
)

// CustomClaims 自定义 claims
type CustomClaims struct {
	jwt.RegisteredClaims
	Audience uint64 `json:"aud,omitempty"` // 为了兼容 easyswoole 会将用户id，设置在 aud 选项
}

// TokenGen 生成
type TokenGen struct {
	issuer  string
	signKey []byte
	nowFunc func() time.Time
}

func NewJwtTokenGen(issuer string, signKey []byte) *TokenGen {
	return &TokenGen{
		issuer:  issuer,
		signKey: signKey,
		nowFunc: time.Now,
	}
}

// GenerateToken 生成 token
func (t *TokenGen) GenerateToken(id uint64, expireSec time.Duration) (string, error) {
	guid := xid.New()
	nowSec := t.nowFunc()
	claims := CustomClaims{}
	claims.ID = guid.String()
	claims.Issuer = t.issuer
	claims.IssuedAt = jwt.NewNumericDate(nowSec)
	claims.ExpiresAt = jwt.NewNumericDate(nowSec.Add(expireSec))
	claims.Audience = id
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtStr, err := token.SignedString(t.signKey)
	if err != nil {
		return "", err
	}

	//return fmt.Sprintf(bearerFormat, jwtStr), nil
	return jwtStr, nil
}

// TokenValidator token 校验
type TokenValidator struct {
	signKey []byte
}

func NewTokenValidator(signKey []byte) *TokenValidator {
	return &TokenValidator{
		signKey: signKey,
	}
}

// Validator 校验 token
func (v *TokenValidator) Validator(token string, options ...jwt.ParserOption) (*CustomClaims, error) {
	/*auths := strings.SplitN(token, " ", 2)
	if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
		return nil, ErrMissingJwtToken
	}
	jwtToken := auths[1]*/
	var (
		tokenInfo *jwt.Token
		err       error
	)
	tokenInfo, err = jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return v.signKey, nil
	}, options...)

	if err != nil {
		// token 无效或过期
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrExpiredOrNotValid
		}
		// 其他错误，如：格式有误 jwt.ErrTokenMalformed、签名错误 jwt.ErrTokenSignatureInvalid、其他错误
		return nil, ErrToken
	}

	if tokenInfo == nil {
		return nil, ErrToken
	}

	if claims, ok := tokenInfo.Claims.(*CustomClaims); ok && tokenInfo.Valid {
		return claims, nil
	}
	return nil, ErrToken
}
