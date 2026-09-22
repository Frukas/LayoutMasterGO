package repository

import (
	"context"
	"errors"
	"fmt"
	"layoutmastergo/internal/models"
	"log/slog"

	"gorm.io/gorm"
)

type productRepository struct {
	db  *gorm.DB
	log *slog.Logger
}

type ProductRepository interface {
	Save(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.Product, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]models.Product, error)
	Update(ctx context.Context, id uint, product *models.Product) error
	Delete(ctx context.Context, id uint) error
}

func NewProductRepository(db *gorm.DB, logger *slog.Logger) ProductRepository {
	return &productRepository{
		db:  db,
		log: logger,
	}
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product

	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Container").
		First(&product, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Warn("product not found", slog.Int("id", int(id)))
			return nil, fmt.Errorf("product with ID %d not found", id)
		}

		r.log.Error("failed to query product in database",
			slog.Int("id", int(id)),
			slog.String("db_error", err.Error()),
		)

		return nil, fmt.Errorf("internal error while fetching product: %w", err)
	}

	return &product, nil
}

func (r *productRepository) Save(ctx context.Context, product *models.Product) error {
	err := r.db.WithContext(ctx).Create(product).Error

	if err != nil {
		r.log.Error("error inserting new product",
			slog.Any("product_payload", product),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to save product: %w", err)
	}

	r.log.Info("product created successfully", slog.Int("product_id", int(product.ID)))
	return nil
}

func (r *productRepository) Update(ctx context.Context, id uint, updatedData *models.Product) error {
	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", id).
		Updates(updatedData).Error

	if err != nil {
		r.log.Error("error updating product",
			slog.Int("product_id", int(id)),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to update product %d: %w", id, err)
	}

	r.log.Info("product updated successfully", slog.Int("product_id", int(id)))
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).
		Delete(&models.Product{}, id).Error

	if err != nil {
		r.log.Error("error performing soft delete on product",
			slog.Int("product_id", int(id)),
			slog.String("db_error", err.Error()),
		)
		return fmt.Errorf("failed to delete product %d: %w", id, err)
	}

	r.log.Info("product deleted successfully", slog.Int("product_id", int(id)))
	return nil
}

// GetAll restaurado para a implementação e assinatura originais
func (r *productRepository) GetAll(ctx context.Context, page, pageSize int) ([]models.Product, error) {
	var products []models.Product

	err := r.db.WithContext(ctx).
		Scopes(Paginate(page, pageSize)).
		Preload("Items").
		Preload("Items.Container").
		Find(&products).Error

	if err != nil {
		r.log.Error("failed to list products", slog.String("db_error", err.Error()))
		return nil, fmt.Errorf("internal error while listing products: %w", err)
	}

	return products, nil
}

// Search método exclusivo para buscas por termo
func (r *productRepository) Search(ctx context.Context, query string, page, pageSize int) ([]models.Product, error) {
	var products []models.Product
	searchTerm := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("name LIKE ? OR jan_code LIKE ?", searchTerm, searchTerm).
		Scopes(Paginate(page, pageSize)).
		Preload("Items").
		Preload("Items.Container").
		Find(&products).Error

	if err != nil {
		r.log.Error("failed to search products", slog.String("db_error", err.Error()), slog.String("query", query))
		return nil, fmt.Errorf("internal error while searching products: %w", err)
	}

	return products, nil
}
