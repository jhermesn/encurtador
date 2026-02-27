package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"encurtador/internal/model"
)

type mysqlURLRepository struct {
	db *sqlx.DB
}

func NewMySQLURLRepository(db *sqlx.DB) URLRepository {
	return &mysqlURLRepository{db: db}
}

func (r *mysqlURLRepository) Create(ctx context.Context, url *model.URL) error {
	query := `
		INSERT INTO urls (slug, target_url, password_hash, manage_token_hash, expires_at)
		VALUES (:slug, :target_url, :password_hash, :manage_token_hash, :expires_at)`
	if _, err := r.db.NamedExecContext(ctx, query, url); err != nil {
		return fmt.Errorf("inserting url: %w", err)
	}
	return nil
}

func (r *mysqlURLRepository) FindBySlug(ctx context.Context, slug string) (*model.URL, error) {
	var url model.URL
	query := `
		SELECT id, slug, target_url, password_hash, manage_token_hash, expires_at, created_at
		FROM urls
		WHERE slug = ? AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &url, query, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding url by slug: %w", err)
	}
	return &url, nil
}

func (r *mysqlURLRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM urls WHERE slug = ?)`, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking slug existence: %w", err)
	}
	return exists, nil
}

func (r *mysqlURLRepository) ExpireBySlug(ctx context.Context, slug, manageTokenHash string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE urls SET expires_at = NOW() WHERE slug = ? AND manage_token_hash = ? AND expires_at > NOW()`,
		slug, manageTokenHash)
	if err != nil {
		return false, fmt.Errorf("expiring url: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

func (r *mysqlURLRepository) DeleteExpired(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM urls WHERE expires_at < NOW()`); err != nil {
		return fmt.Errorf("deleting expired urls: %w", err)
	}
	return nil
}
