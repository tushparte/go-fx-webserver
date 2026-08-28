package user

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users (id, name) VALUES (?, ?)", u.ID, u.Name)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
