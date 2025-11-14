package repository

import (
	"context"
	"fmt"

	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/jmoiron/sqlx"
)

type refreshSessionRepository struct {
	db *sqlx.DB
}

func newRefreshSessionRepository(db *sqlx.DB) *refreshSessionRepository {
	return &refreshSessionRepository{
		db: db,
	}
}

func (r *refreshSessionRepository) Create(ctx context.Context, session *domain.RefreshSession) error {
	const query = `
				INSERT INTO refresh_session (id, user_id, refresh_token, user_agent, ip, expires_in)
				VALUES (uuid_to_bin(?), uuid_to_bin(?), uuid_to_bin(?), ?, ?, ?)
				`
	_, err := r.db.ExecContext(ctx, query, session.ID, session.UserID, session.RefreshToken, session.UserAgent, session.IP, session.ExpiresIn)

	if err != nil {
		return fmt.Errorf("db insert user: %w", err)
	}

	return nil

}
