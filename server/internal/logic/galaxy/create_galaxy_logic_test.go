//go:build integration

package galaxy

import (
	"context"
	"testing"

	"github.com/stellaris/stellaris/server/internal/model"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"
)

func TestCreateGalaxy(t *testing.T) {
	db := model.OpenTestDB(t)
	sc := &svc.ServiceContext{
		DB:       db,
		Users:    model.NewUserModel(db),
		Galaxies: model.NewGalaxyModel(db),
	}
	sc.Config.UserJwt.AccessSecret = "secret"
	sc.Config.UserJwt.AccessExpire = 3600

	uid, err := sc.Users.Create(context.Background(), "u1@x.com", "hash")
	if err != nil {
		t.Fatal(err)
	}

	// 模拟 JWT 中间件注入的 id claim（float64 是 go-zero 的常态）
	ctx := context.WithValue(context.Background(), "id", float64(uid))

	resp, err := NewCreateGalaxyLogic(ctx, sc).CreateGalaxy(&types.GalaxyCreateReq{Name: "Andromeda"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.GID) != 16 {
		t.Fatalf("gid len=%d", len(resp.GID))
	}
	if len(resp.NodeToken) != 64 {
		t.Fatalf("node_token len=%d", len(resp.NodeToken))
	}

	g, err := sc.Galaxies.FindByGID(context.Background(), resp.GID)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "Andromeda" || g.OwnerUserID != uid {
		t.Fatalf("galaxy mismatch: %+v", g)
	}
}
