package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type UserDocumentRepository interface {
	Create(ctx context.Context, document *domain.UserDocument) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserDocument, error)
	Update(ctx context.Context, document *domain.UserDocument) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userDocumentRepository struct {
	db *sqlx.DB
}

func NewUserDocumentRepository(db *sqlx.DB) UserDocumentRepository {
	return &userDocumentRepository{
		db: db,
	}
}

func (r *userDocumentRepository) Create(ctx context.Context, document *domain.UserDocument) error {
	const query = `
		INSERT INTO user_document (id, user_id, document_type, document_number, created_at, updated_at, deleted_at)
		VALUES (uuid_to_bin(?), uuid_to_bin(?), ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, document.ID, document.UserID, document.DocumentType, document.DocumentNumber, document.CreatedAt, document.UpdatedAt, document.DeletedAt)
	if err != nil {
		return fmt.Errorf("db insert user document: %w", err)
	}
	return nil
}

func (r *userDocumentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserDocument, error) {
	const query = `
		SELECT id, user_id, document_type, document_number, created_at, updated_at, deleted_at FROM user_document WHERE user_id = uuid_to_bin(?)
	`
	var documents []domain.UserDocument
	err := r.db.SelectContext(ctx, &documents, query, userID)
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func (r *userDocumentRepository) Update(ctx context.Context, document *domain.UserDocument) error {
	const query = `
		UPDATE user_document SET document_type = ?, document_number = ?, updated_at = ?, deleted_at = ? WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, document.DocumentType, document.DocumentNumber, document.UpdatedAt, document.DeletedAt, document.ID)
	if err != nil {
		return fmt.Errorf("db update user document: %w", err)
	}
	return nil
}

func (r *userDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE user_document SET deleted_at = NOW() WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db delete user document: %w", err)
	}
	return nil
}
