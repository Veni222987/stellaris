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

// Upsert 用于 Planet 首次 join 或重新上线时落库。
func (m *PlanetModel) Upsert(ctx context.Context, planetUUID, gid, ip, hostname, osName string) (int64, error) {
	var id int64
	err := m.db.QueryRow(ctx, `
		INSERT INTO planets (planet_uuid, gid, ip, hostname, os, status, last_heartbeat)
		VALUES ($1, $2, $3, $4, $5, 'online', NOW())
		ON CONFLICT (planet_uuid) DO UPDATE SET
			ip = EXCLUDED.ip,
			hostname = EXCLUDED.hostname,
			os = EXCLUDED.os,
			status = 'online',
			last_heartbeat = NOW()
		RETURNING id`,
		planetUUID, gid, ip, hostname, osName).Scan(&id)
	return id, err
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
