package model

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Galaxy struct {
	ID            int64
	GID           string
	Name          string
	NodeTokenHash string
	OwnerUserID   int64
}

type GalaxyModel struct{ db *pgxpool.Pool }

func NewGalaxyModel(db *pgxpool.Pool) *GalaxyModel { return &GalaxyModel{db: db} }

func (m *GalaxyModel) Create(ctx context.Context, gid, name, nodeTokenHash string, ownerUserID int64) (int64, error) {
	var id int64
	err := m.db.QueryRow(ctx,
		`INSERT INTO galaxies (gid, name, node_token_hash, owner_user_id)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		gid, name, nodeTokenHash, ownerUserID).Scan(&id)
	return id, err
}

func (m *GalaxyModel) FindByGID(ctx context.Context, gid string) (*Galaxy, error) {
	var g Galaxy
	err := m.db.QueryRow(ctx,
		`SELECT id, gid, name, node_token_hash, owner_user_id FROM galaxies WHERE gid = $1`,
		gid).Scan(&g.ID, &g.GID, &g.Name, &g.NodeTokenHash, &g.OwnerUserID)
	return &g, err
}
