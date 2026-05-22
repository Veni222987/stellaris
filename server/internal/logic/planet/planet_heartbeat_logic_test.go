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

func TestPlanetHeartbeat(t *testing.T) {
	db := model.OpenTestDB(t)
	sc := &svc.ServiceContext{
		DB:       db,
		Users:    model.NewUserModel(db),
		Galaxies: model.NewGalaxyModel(db),
		Planets:  model.NewPlanetModel(db),
		Agents:   model.NewAgentModel(db),
		Config:   config.Config{},
	}

	ctx := context.Background()
	uid, _ := sc.Users.Create(ctx, "u@x.com", "x")
	h, _ := bcrypt.GenerateFromPassword([]byte("tok"), bcrypt.DefaultCost)
	if _, err := sc.Galaxies.Create(ctx, "g1", "G", string(h), uid); err != nil {
		t.Fatal(err)
	}
	pid, err := sc.Planets.Upsert(ctx, "puuid", "g1", "1.1.1.1", "h", "linux")
	if err != nil {
		t.Fatal(err)
	}

	// 模拟 JWT 注入 planet_id（go-zero 实际是 float64）
	jwtCtx := context.WithValue(ctx, "planet_id", float64(pid))

	resp, err := NewPlanetHeartbeatLogic(jwtCtx, sc).PlanetHeartbeat(&types.HeartbeatReq{
		Agents: []types.AgentInfo{
			{
				AgentUUID:        "ag-uuid-1",
				Type:             "openclaw",
				Name:             "claw-1",
				Models:           []string{"default"},
				CapabilitiesJSON: map[string]interface{}{"streaming": true},
				Status:           "ready",
			},
			{
				AgentUUID:        "ag-uuid-2",
				Type:             "hermes",
				Name:             "hermes-1",
				Models:           []string{"default"},
				CapabilitiesJSON: map[string]interface{}{},
				Status:           "ready",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatal("expected OK=true")
	}

	a, err := sc.Agents.FindByUUID(ctx, "ag-uuid-1")
	if err != nil {
		t.Fatal(err)
	}
	if a.Type != "openclaw" || a.PlanetID != pid {
		t.Fatalf("agent mismatch: %+v", a)
	}
}
