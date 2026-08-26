package main

import (
	"context"
	"database/sql"
)

type User struct {
	ID   string
	Name string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *User) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, name) VALUES (? , ?)", u.ID, u.Name)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
