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
	ErrInvalidQuantity    = errors.New("A quantidade deve ser maior que zero.")
	ErrItemNotFound       = errors.New("Produto não encontrado neste contêiner.")
	ErrInvalidStatus      = errors.New("Status do contêiner inválido.")
	ErrContainerBlankName = errors.New("O nome do contêiner não pode ficar em branco.")
	ErrContainerDuplicate = errors.New("Já existe um contêiner cadastrado com este Nome.")
	ErrProductNotFound    = errors.New("O produto selecionado não existe.")
	ErrContainerNotFound  = errors.New("O contêiner informado não foi encontrado.")
)

type ContainerService interface {
	Create(ctx context.Context, container *models.Container) error
	GetByID(ctx context.Context, id uint) (*models.Container, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.Container, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]models.Container, error)
	Update(ctx context.Context, id uint, updateData *models.Container) error
	Delete(ctx context.Context, id uint) error

	AddProduct(ctx context.Context, containerID, productID uint, quantity int) error
	UpdateProductQuantity(ctx context.Context, containerID, productID uint, quantity int) error
	RemoveProduct(ctx context.Context, containerID, productID uint) error
}

type containerService struct {
	containerRepo     repository.ContainerRepository
	itemContainerRepo repository.ItemContainerRepository
	productRepo       repository.ProductRepository
	log               *slog.Logger
}

func NewContainerService(
	containerRepo repository.ContainerRepository,
	itemContainerRepo repository.ItemContainerRepository,
	productRepo repository.ProductRepository,
	logger *slog.Logger,
) ContainerService {
	return &containerService{
		containerRepo:     containerRepo,
		itemContainerRepo: itemContainerRepo,
		productRepo:       productRepo,
		log:               logger,
	}
}

func (s *containerService) Create(ctx context.Context, container *models.Container) error {
	container.Name = strings.TrimSpace(container.Name)
	if container.Name == "" {
		return ErrContainerBlankName
	}
	if container.Status == "" {
		container.Status = "Pending"
	}

	if err := s.containerRepo.Save(ctx, container); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return ErrContainerDuplicate
		}
		return fmt.Errorf("falha ao criar contêiner: %w", err)
	}
	return nil
}

func (s *containerService) GetByID(ctx context.Context, id uint) (*models.Container, error) {
	return s.containerRepo.GetByID(ctx, id)
}

func (s *containerService) GetAll(ctx context.Context, page, pageSize int) ([]models.Container, error) {
	return s.containerRepo.GetAll(ctx, page, pageSize)
}

func (s *containerService) Search(ctx context.Context, query string, page, pageSize int) ([]models.Container, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.GetAll(ctx, page, pageSize)
	}
	return s.containerRepo.Search(ctx, query, page, pageSize)
}

func (s *containerService) Update(ctx context.Context, id uint, updateData *models.Container) error {
	if updateData.Name != "" {
		updateData.Name = strings.TrimSpace(updateData.Name)
		if updateData.Name == "" {
			return ErrContainerBlankName
		}
	}

	if updateData.Status != "" && updateData.Status != "Pending" && updateData.Status != "Active" {
		s.log.Warn("attempt to update container with invalid status", slog.String("status", updateData.Status))
		return ErrInvalidStatus
	}

	if err := s.containerRepo.Update(ctx, id, updateData); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return ErrContainerDuplicate
		}
		return fmt.Errorf("falha ao atualizar contêiner: %w", err)
	}
	return nil
}

func (s *containerService) Delete(ctx context.Context, id uint) error {
	if err := s.containerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("falha ao excluir contêiner: %w", err)
	}
	return nil
}

func (s *containerService) AddProduct(ctx context.Context, containerID, productID uint, quantity int) error {
	s.log.Info("processing AddProduct request", slog.Int("container_id", int(containerID)), slog.Int("product_id", int(productID)))

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return ErrProductNotFound
	}

	container, err := s.containerRepo.GetByID(ctx, containerID)
	if err != nil {
		return ErrContainerNotFound
	}

	for _, item := range container.Items {
		if item.ProductID == productID {
			newQuantity := item.Quantity + quantity
			s.log.Info("product already exists in container, updating quantity instead",
				slog.Int("old_qty", item.Quantity),
				slog.Int("new_qty", newQuantity))

			return s.itemContainerRepo.UpdateQuantity(ctx, containerID, productID, newQuantity)
		}
	}

	newItem := &models.ItemContainer{
		ContainerID: containerID,
		ProductID:   productID,
		Quantity:    quantity,
	}

	if err := s.itemContainerRepo.AddItem(ctx, newItem); err != nil {
		return fmt.Errorf("falha ao adicionar produto ao contêiner: %w", err)
	}
	return nil
}

func (s *containerService) UpdateProductQuantity(ctx context.Context, containerID, productID uint, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	container, err := s.containerRepo.GetByID(ctx, containerID)
	if err != nil {
		return ErrContainerNotFound
	}

	exists := false
	for _, item := range container.Items {
		if item.ProductID == productID {
			exists = true
			break
		}
	}

	if !exists {
		return ErrItemNotFound
	}

	if err := s.itemContainerRepo.UpdateQuantity(ctx, containerID, productID, quantity); err != nil {
		return fmt.Errorf("falha ao atualizar a quantidade do produto: %w", err)
	}
	return nil
}

func (s *containerService) RemoveProduct(ctx context.Context, containerID, productID uint) error {
	if err := s.itemContainerRepo.RemoveItem(ctx, containerID, productID); err != nil {
		return fmt.Errorf("falha ao remover produto do contêiner: %w", err)
	}

	s.log.Info("product removed from container successfully", slog.Int("container_id", int(containerID)), slog.Int("product_id", int(productID)))
	return nil
}
