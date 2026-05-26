package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Planet struct {
	ID         int64
	PlanetUUID string
	GID        string
	IP         string
	Hostname   string
	OS         string
	Status     string
}

type PlanetModel struct{ db *pgxpool.Pool }

func NewPlanetModel(db *pgxpool.Pool) *PlanetModel { return &PlanetModel{db: db} }

// Upsert 按 (gid, hostname) 去重：同一主机重新 orbit 时复用已有行，UUID 保持稳定。
// 返回 (id, actualPlanetUUID, error)；若行已存在则 actualPlanetUUID 是原有值。
func (m *PlanetModel) Upsert(ctx context.Context, planetUUID, gid, ip, hostname, osName string) (int64, string, error) {
	var id int64
	var actualUUID string
	err := m.db.QueryRow(ctx, `
		INSERT INTO planets (planet_uuid, gid, ip, hostname, os, status, last_heartbeat)
		VALUES ($1, $2, $3, $4, $5, 'online', NOW())
		ON CONFLICT (gid, hostname) DO UPDATE SET
			ip = EXCLUDED.ip,
			os = EXCLUDED.os,
			status = 'online',
			last_heartbeat = NOW()
		RETURNING id, planet_uuid`,
		planetUUID, gid, ip, hostname, osName).Scan(&id, &actualUUID)
	return id, actualUUID, err
}

func (m *PlanetModel) FindByID(ctx context.Context, id int64) (*Planet, error) {
	var p Planet
	err := m.db.QueryRow(ctx,
		`SELECT id, planet_uuid, gid, ip, hostname, COALESCE(os, ''), status FROM planets WHERE id = $1`,
		id).Scan(&p.ID, &p.PlanetUUID, &p.GID, &p.IP, &p.Hostname, &p.OS, &p.Status)
	return &p, err
}

// UpdateHeartbeat 在收到心跳时刷新 last_heartbeat 与 status。
func (m *PlanetModel) UpdateHeartbeat(ctx context.Context, planetID int64) error {
	_, err := m.db.Exec(ctx,
		`UPDATE planets SET last_heartbeat = NOW(), status = 'online' WHERE id = $1`,
		planetID)
	return err
}
