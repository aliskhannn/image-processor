package image

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/aliskhannn/image-processor/internal/model"
)

var ErrImageNotFound = errors.New("image not found")

type Repository struct {
	db *dbpg.DB
}

func NewRepository(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveImage(ctx context.Context, img model.Image) (uuid.UUID, error) {
	query := `
		INSERT INTO images (original_id, filename, path, action, params, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
   `

	var id uuid.UUID
	err := r.db.Master.QueryRowContext(
		ctx, query, img.OriginalID, img.Filename, img.Path, img.Action.Name, img.Action.Params, img.Status,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save: failed to save image: %w", err)
	}

	return id, nil
}

func (r *Repository) GetImage(ctx context.Context, id uuid.UUID) (model.Image, error) {
	query := `
		SELECT original_id, filename, path, action, params, status, created_at
		FROM images
		WHERE id = $1
    `

	var img model.Image
	img.ID = id
	err := r.db.Master.QueryRowContext(
		ctx, query, id,
	).Scan(&img.OriginalID, &img.Filename, &img.Path, &img.Action.Name, &img.Action.Params, &img.Status, &img.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Image{}, ErrImageNotFound
		}

		return model.Image{}, fmt.Errorf("get: failed to get image: %w", err)
	}

	return img, nil
}

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
