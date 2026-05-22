// Package jwtctx 从 go-zero JWT 中间件注入的 context 里安全地提取 claim 值。
//
// go-zero 把 JWT payload 的每个 key 直接 put 进 context，value 的实际 Go 类型依赖底层 JSON 解析行为：
// 数字通常是 float64，但启用 UseNumber 时是 json.Number。本包统一兜底。
package jwtctx

import (
	"context"
	"encoding/json"
	"errors"
)

// UserID 取 user JWT 里的 id claim。
func UserID(ctx context.Context) (int64, error) {
	return intClaim(ctx, "id")
}

// PlanetID 取 planet JWT 里的 planet_id claim（兼容回退到 id）。
func PlanetID(ctx context.Context) (int64, error) {
	if v, err := intClaim(ctx, "planet_id"); err == nil {
		return v, nil
	}
	return intClaim(ctx, "id")
}

// GID 取 planet JWT 里的 gid claim。
func GID(ctx context.Context) (string, error) {
	v := ctx.Value("gid")
	s, ok := v.(string)
	if !ok || s == "" {
		return "", errors.New("missing gid claim")
	}
	return s, nil
}

func intClaim(ctx context.Context, key string) (int64, error) {
	v := ctx.Value(key)
	switch x := v.(type) {
	case float64:
		return int64(x), nil
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	case json.Number:
		return x.Int64()
	case string:
		// 极端兜底：如果对端把数字编成字符串
		var n json.Number = json.Number(x)
		return n.Int64()
	}
	return 0, errors.New("claim " + key + " missing or unsupported type")
}
