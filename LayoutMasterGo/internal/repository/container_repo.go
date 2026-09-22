package repository

import (
	"context"
	"errors"
	"fmt"
	"layoutmastergo/internal/models"
	"log/slog"

	"gorm.io/gorm"
)

type containerRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

type ContainerRepository interface {
	Save(ctx context.Context, container *models.Container) error
	GetByID(ctx context.Context, id uint) (*models.Container, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.Container, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]models.Container, error) // Novo método de busca
	Update(ctx context.Context, id uint, container *models.Container) error
	Delete(ctx context.Context, id uint) error
}

func NewContainerRepository(db *gorm.DB, logger *slog.Logger) ContainerRepository {
	return &containerRepository{
		db:  db,
		log: logger,
	}
}

func (r *containerRepository) GetByID(ctx context.Context, id uint) (*models.Container, error) {
	var container models.Container

	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		First(&container, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Warn("container not found", slog.Int("id", int(id)))
			return nil, fmt.Errorf("container with ID %d not found", id)
		}
		r.log.Error("failed to query container in database", slog.Int("id", int(id)), slog.String("db_error", err.Error()))
		return nil, fmt.Errorf("internal error while fetching container: %w", err)
	}

	return &container, nil
}

func (r *containerRepository) Save(ctx context.Context, container *models.Container) error {
	err := r.db.WithContext(ctx).Create(container).Error
	if err != nil {
		r.log.Error("error inserting new container and items", slog.Any("container_payload", container), slog.String("db_error", err.Error()))
		return fmt.Errorf("failed to save container: %w", err)
	}
	r.log.Info("container created successfully", slog.Int("container_id", int(container.ID)))
	return nil
}

func (r *containerRepository) Update(ctx context.Context, id uint, updatedData *models.Container) error {
	err := r.db.WithContext(ctx).
		Model(&models.Container{}).
		Where("id = ?", id).
		Updates(updatedData).Error

	if err != nil {
		r.log.Error("error updating container", slog.Int("container_id", int(id)), slog.String("db_error", err.Error()))
		return fmt.Errorf("failed to update container %d: %w", id, err)
	}
	r.log.Info("container updated successfully", slog.Int("container_id", int(id)))
	return nil
}

func (r *containerRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&models.Container{}, id).Error
	if err != nil {
		r.log.Error("error performing soft delete on container", slog.Int("container_id", int(id)), slog.String("db_error", err.Error()))
		return fmt.Errorf("failed to delete container %d: %w", id, err)
	}
	r.log.Info("container deleted successfully", slog.Int("container_id", int(id)))
	return nil
}

func (r *containerRepository) GetAll(ctx context.Context, page, pageSize int) ([]models.Container, error) {
	var containers []models.Container

	err := r.db.WithContext(ctx).
		Scopes(Paginate(page, pageSize)).
		Preload("Items").
		Preload("Items.Product").
		Find(&containers).Error

	if err != nil {
		r.log.Error("failed to list containers", slog.String("db_error", err.Error()))
		return nil, fmt.Errorf("internal error while listing containers: %w", err)
	}

	return containers, nil
}

// Search isolado para regras de busca (Nome ou Status)
func (r *containerRepository) Search(ctx context.Context, query string, page, pageSize int) ([]models.Container, error) {
	var containers []models.Container
	searchTerm := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("name LIKE ? OR status LIKE ?", searchTerm, searchTerm).
		Scopes(Paginate(page, pageSize)).
		Preload("Items").
		Preload("Items.Product").
		Find(&containers).Error

	if err != nil {
		r.log.Error("failed to search containers", slog.String("db_error", err.Error()), slog.String("query", query))
		return nil, fmt.Errorf("internal error while searching containers: %w", err)
	}

	return containers, nil
}
