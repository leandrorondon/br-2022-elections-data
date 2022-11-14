package steps

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Exists(ctx context.Context, step string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT * FROM steps_completed WHERE step = $1)`
	err := r.db.GetContext(ctx, &exists, query, step)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return exists, nil
}

func (r *Repository) Insert(ctx context.Context, step string) error {
	query := `INSERT INTO steps_completed(step) VALUES($1) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, step)
	return err
}
