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

func (m *GalaxyModel) ListByUser(ctx context.Context, ownerUserID int64) ([]Galaxy, error) {
	rows, err := m.db.Query(ctx,
		`SELECT id, gid, name, node_token_hash, owner_user_id FROM galaxies
		 WHERE owner_user_id = $1 ORDER BY id DESC`,
		ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Galaxy
	for rows.Next() {
		var g Galaxy
		if err := rows.Scan(&g.ID, &g.GID, &g.Name, &g.NodeTokenHash, &g.OwnerUserID); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (m *GalaxyModel) FindByGID(ctx context.Context, gid string) (*Galaxy, error) {
	var g Galaxy
	err := m.db.QueryRow(ctx,
		`SELECT id, gid, name, node_token_hash, owner_user_id FROM galaxies WHERE gid = $1`,
		gid).Scan(&g.ID, &g.GID, &g.Name, &g.NodeTokenHash, &g.OwnerUserID)
	return &g, err
}
