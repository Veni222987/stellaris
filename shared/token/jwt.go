// Package token 提供 Stellaris 内部 JWT 颁发与解析的最小工具。
//
// 两类主体共用同一套 Claims 结构：
//   - 面板用户：Sub=user, ID=user_id
//   - Planet 节点：Sub=planet, ID=planet_id, GID=gid, PlanetID=planet_id
//
// AccessSecret 来自 server/etc/core.yaml 中的 UserJwt/PlanetJwt 配置。
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Subject string

const (
	SubjectUser   Subject = "user"
	SubjectPlanet Subject = "planet"
)

type Claims struct {
	Sub      Subject `json:"sub"`
	ID       int64   `json:"id"`
	GID      string  `json:"gid,omitempty"`
	PlanetID int64   `json:"planet_id,omitempty"`
	jwt.RegisteredClaims
}

// Issue 颁发一个 HS256 签名的 JWT。expireSec 为过期秒数。
func Issue(secret string, expireSec int64, c Claims) (string, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(expireSec) * time.Second))
	c.IssuedAt = jwt.NewNumericDate(time.Now())
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(secret))
}

// Parse 校验并解析 token，返回 Claims 或错误。
func Parse(secret, raw string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("token invalid")
	}
	return c, nil
}
