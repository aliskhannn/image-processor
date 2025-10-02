package image

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/image-processor/internal/model"
)

var ErrImageNotFound = errors.New("image not found")

// Repository provides CRUD operations for images in the database.
type Repository struct {
	db *dbpg.DB
}

// NewRepository creates a new Repository with the given DB connection.
func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

// SaveImage inserts a new image record into the database and returns its UUID.
func (r *Repository) SaveImage(ctx context.Context, img model.Image) (uuid.UUID, error) {
	query := `
		INSERT INTO images (filename, path, action, params, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
   `

	paramsJSON, err := json.Marshal(img.Action.Params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal action params: %w", err)
	}

	var id uuid.UUID
	err = r.db.QueryRowContext(
		ctx, query, img.Filename, img.Path, img.Action.Name, paramsJSON, img.Status,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save: failed to save image: %w", err)
	}

	return id, nil
}

// GetImage retrieves an image record by ID from the database.
func (r *Repository) GetImage(ctx context.Context, id uuid.UUID) (model.Image, error) {
	query := `
		SELECT filename, path, action, params, status, created_at
		FROM images
		WHERE id = $1
    `

	var img model.Image
	var paramsBytes []byte

	err := r.db.QueryRowContext(
		ctx, query, id,
	).Scan(&img.Filename, &img.Path, &img.Action.Name, &paramsBytes, &img.Status, &img.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Image{}, ErrImageNotFound
		}

		return model.Image{}, fmt.Errorf("get: failed to get image: %w", err)
	}

	if err := json.Unmarshal(paramsBytes, &img.Action.Params); err != nil {
		return model.Image{}, fmt.Errorf("get: failed to unmarshal params: %w", err)
	}

	img.ID = id

	return img, nil
}

// UpdateImage updates the path and status of an existing image by ID.
func (r *Repository) UpdateImage(ctx context.Context, id uuid.UUID, path, status string) error {
	query := `
		UPDATE images
		SET path = $1, status = $2
		WHERE id = $3
    `

	res, err := r.db.ExecContext(ctx, query, path, status, id)
	if err != nil {
		return fmt.Errorf("update: failed to update image: %w", err)
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return ErrImageNotFound
	}

	return nil
}

// DeleteImage deletes an image record by ID from the database.
func (r *Repository) DeleteImage(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM images WHERE id = $1
    `

	rows, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete: failed to delete image: %w", err)
	}

	n, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete: failed to get number of rows affected: %w", err)
	}

	if n == 0 {
		return ErrImageNotFound
	}

	return nil
}
