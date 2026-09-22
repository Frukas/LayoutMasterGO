package repository

import (
	"context"
	"fmt"
	"layoutmastergo/internal/models"
	"log/slog"

	"gorm.io/gorm"
)

type itemContainerRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

// ItemContainerRepository handles surgical operations specifically on the join table
type ItemContainerRepository interface {
	AddItem(ctx context.Context, item *models.ItemContainer) error
	UpdateQuantity(ctx context.Context, containerID, productID uint, quantity int) error
	RemoveItem(ctx context.Context, containerID, productID uint) error
}

func NewItemContainerRepository(db *gorm.DB, logger *slog.Logger) ItemContainerRepository {
	return &itemContainerRepository{
		db:  db,
		log: logger,
	}
}

// AddItem inserts a new relation between a container and a product
func (r *itemContainerRepository) AddItem(ctx context.Context, item *models.ItemContainer) error {
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		r.log.Error("error adding item to container",
			slog.Int("container_id", int(item.ContainerID)),
			slog.Int("product_id", int(item.ProductID)),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to add item to container: %w", err)
	}
	return nil
}

// UpdateQuantity updates ONLY the quantity of an existing relation
func (r *itemContainerRepository) UpdateQuantity(ctx context.Context, containerID, productID uint, quantity int) error {
	// We use Model and Where with composite keys to update a specific column directly
	err := r.db.WithContext(ctx).
		Model(&models.ItemContainer{}).
		Where("container_id = ? AND product_id = ?", containerID, productID).
		Update("quantity", quantity).Error

	if err != nil {
		r.log.Error("error updating item quantity",
			slog.Int("container_id", int(containerID)),
			slog.Int("product_id", int(productID)),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to update item quantity: %w", err)
	}
	return nil
}

// RemoveItem hard deletes the relation (the product is no longer in this container)
func (r *itemContainerRepository) RemoveItem(ctx context.Context, containerID, productID uint) error {
	// For join tables, we often use hard delete to simply sever the link,
	// unless you need audit history on the join table itself.
	err := r.db.WithContext(ctx).
		Where("container_id = ? AND product_id = ?", containerID, productID).
		Delete(&models.ItemContainer{}).Error

	if err != nil {
		r.log.Error("error removing item from container",
			slog.Int("container_id", int(containerID)),
			slog.Int("product_id", int(productID)),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to remove item from container: %w", err)
	}
	return nil
}
