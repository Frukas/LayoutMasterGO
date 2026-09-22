package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"layoutmastergo/internal/models"
	"layoutmastergo/internal/repository"
)

var (
	ErrInvalidProductName = errors.New("O nome do produto não pode ficar em branco.")
	ErrInvalidJanCode     = errors.New("O código JAN do produto não pode ficar em branco.")
	ErrProductDuplicate   = errors.New("Já existe um produto cadastrado com este Nome ou JanCode.")
	ErrProductInUse       = errors.New("Não é possível excluir este produto pois ele está vinculado a um ou mais contêineres.")
)

type ProductService interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.Product, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]models.Product, error)
	Update(ctx context.Context, id uint, updateData *models.Product) error
	Delete(ctx context.Context, id uint) error
}

type productService struct {
	productRepo repository.ProductRepository
	log         *slog.Logger
}

func NewProductService(productRepo repository.ProductRepository, logger *slog.Logger) ProductService {
	return &productService{
		productRepo: productRepo,
		log:         logger,
	}
}

func (s *productService) Create(ctx context.Context, product *models.Product) error {
	product.Name = strings.TrimSpace(product.Name)
	product.JanCode = strings.TrimSpace(product.JanCode)

	if product.Name == "" {
		s.log.Warn("attempt to create product with invalid name")
		return ErrInvalidProductName
	}
	if product.JanCode == "" {
		s.log.Warn("attempt to create product with invalid JAN code")
		return ErrInvalidJanCode
	}

	if err := s.productRepo.Save(ctx, product); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return ErrProductDuplicate
		}
		return fmt.Errorf("falha ao criar produto: %w", err)
	}

	return nil
}

func (s *productService) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *productService) GetAll(ctx context.Context, page, pageSize int) ([]models.Product, error) {
	return s.productRepo.GetAll(ctx, page, pageSize)
}

func (s *productService) Search(ctx context.Context, query string, page, pageSize int) ([]models.Product, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.GetAll(ctx, page, pageSize)
	}
	return s.productRepo.Search(ctx, query, page, pageSize)
}

func (s *productService) Update(ctx context.Context, id uint, updateData *models.Product) error {
	if updateData.Name != "" {
		updateData.Name = strings.TrimSpace(updateData.Name)
		if updateData.Name == "" {
			s.log.Warn("attempt to update product with blank name", slog.Int("product_id", int(id)))
			return ErrInvalidProductName
		}
	}

	if updateData.JanCode != "" {
		updateData.JanCode = strings.TrimSpace(updateData.JanCode)
		if updateData.JanCode == "" {
			s.log.Warn("attempt to update product with blank JAN code", slog.Int("product_id", int(id)))
			return ErrInvalidJanCode
		}
	}

	if err := s.productRepo.Update(ctx, id, updateData); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return ErrProductDuplicate
		}
		return fmt.Errorf("falha ao atualizar produto: %w", err)
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "foreign") || strings.Contains(errStr, "constraint") || strings.Contains(errStr, "restrict") {
			return ErrProductInUse
		}
		return fmt.Errorf("falha ao excluir produto: %w", err)
	}
	return nil
}
