package planet

import (
	"context"
	"testing"

	"github.com/stellaris/stellaris/server/internal/config"
	"github.com/stellaris/stellaris/server/internal/model"
	"github.com/stellaris/stellaris/server/internal/svc"
	"github.com/stellaris/stellaris/server/internal/types"

	"golang.org/x/crypto/bcrypt"
)

func TestPlanetJoin(t *testing.T) {
	db := model.OpenTestDB(t)
	sc := &svc.ServiceContext{
		DB:       db,
		Users:    model.NewUserModel(db),
		Galaxies: model.NewGalaxyModel(db),
		Planets:  model.NewPlanetModel(db),
		Config:   config.Config{},
	}
	sc.Config.PlanetJwt.AccessSecret = "ps"
	sc.Config.PlanetJwt.AccessExpire = 3600

	ctx := context.Background()
	uid, _ := sc.Users.Create(ctx, "owner@x.com", "x")
	rawToken := "raw-token-1234567890abcdef"
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawToken), bcrypt.DefaultCost)
	if _, err := sc.Galaxies.Create(ctx, "ab12cd34ef567890", "G1", string(hash), uid); err != nil {
		t.Fatal(err)
	}

	resp, err := NewPlanetJoinLogic(ctx, sc).PlanetJoin(&types.PlanetJoinReq{
		GID: "ab12cd34ef567890", NodeToken: rawToken,
		Hostname: "h1", IP: "10.0.0.1", OS: "linux/arm64",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.PlanetUUID) != 32 {
		t.Fatalf("planet_uuid len=%d", len(resp.PlanetUUID))
	}
	if resp.PlanetJwt == "" {
		t.Fatal("empty planet_jwt")
	}

	// 错误 token 应该拒绝
	if _, err := NewPlanetJoinLogic(ctx, sc).PlanetJoin(&types.PlanetJoinReq{
		GID: "ab12cd34ef567890", NodeToken: "wrong",
		Hostname: "h1", IP: "10.0.0.1", OS: "linux",
	}); err == nil {
		t.Fatal("expected error for wrong token")
	}
}
