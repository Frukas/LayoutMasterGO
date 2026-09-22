package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"layoutmastergo/internal/mocks"
	"layoutmastergo/internal/models"
	"layoutmastergo/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupService helper func returns a fresh service and fresh mocks for each test
func setupService(t *testing.T) (
	service.ContainerService,
	*mocks.ContainerRepository,
	*mocks.ItemContainerRepository,
	*mocks.ProductRepository,
) {
	// O Mockery já amarra o 't' no mock, então ele faz o AssertExpectations() automaticamente no final!
	cRepo := mocks.NewContainerRepository(t)
	iRepo := mocks.NewItemContainerRepository(t)
	pRepo := mocks.NewProductRepository(t)

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewContainerService(cRepo, iRepo, pRepo, silentLogger)

	return svc, cRepo, iRepo, pRepo
}

func TestContainerService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path - Fallback to Pending status", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)

		container := &models.Container{Name: "Box A", Status: ""}

		// O mock.MatchedBy permite checar se o serviço alterou o dado (Regra de Negócio) antes de chamar o repo
		cRepo.On("Save", mock.Anything, mock.MatchedBy(func(c *models.Container) bool {
			return c.Status == "Pending"
		})).Return(nil).Once()

		err := svc.Create(ctx, container)
		assert.NoError(t, err)
	})

	t.Run("Error Path - Repo Failure", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)

		container := &models.Container{Name: "Box B", Status: "Active"}
		expectedErr := errors.New("db connection failed")

		cRepo.On("Save", mock.Anything, container).Return(expectedErr).Once()

		err := svc.Create(ctx, container)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr) // Verifica se o erro do db foi "empacotado" corretamente
	})
}

func TestContainerService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Negative Path - Invalid Status", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		updateData := &models.Container{Status: "Completed"}

		err := svc.Update(ctx, 1, updateData)
		assert.ErrorIs(t, err, service.ErrInvalidStatus)
		// Nota: Não configuramos nenhum .On() no mock porque ele DEVE falhar antes de chamar o DB
	})

	t.Run("Happy Path - Valid Status", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)
		updateData := &models.Container{Status: "Active"}

		cRepo.On("Update", mock.Anything, uint(1), updateData).Return(nil).Once()

		err := svc.Update(ctx, 1, updateData)
		assert.NoError(t, err)
	})
}

func TestContainerService_AddProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("Negative Path - Invalid Quantity", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		err := svc.AddProduct(ctx, 1, 1, 0)
		assert.ErrorIs(t, err, service.ErrInvalidQuantity)
	})

	t.Run("Negative Path - Product Does Not Exist", func(t *testing.T) {
		svc, _, _, pRepo := setupService(t)

		pRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("not found")).Once()

		err := svc.AddProduct(ctx, 1, 999, 10)
		assert.Error(t, err)
	})

	t.Run("Happy Path - Product Already in Container (Sum Quantity)", func(t *testing.T) {
		svc, cRepo, iRepo, pRepo := setupService(t)

		pRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.Product{}, nil).Once()

		existingContainer := &models.Container{
			Items: []models.ItemContainer{{ProductID: 1, Quantity: 10}}, // Já tem 10
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(existingContainer, nil).Once()

		// A regra de negócio diz que 10 (existente) + 5 (novo) = 15.
		iRepo.On("UpdateQuantity", mock.Anything, uint(1), uint(1), 15).Return(nil).Once()

		err := svc.AddProduct(ctx, 1, 1, 5)
		assert.NoError(t, err)
	})

	t.Run("Happy Path - Brand New Product in Container", func(t *testing.T) {
		svc, cRepo, iRepo, pRepo := setupService(t)

		pRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.Product{}, nil).Once()

		emptyContainer := &models.Container{Items: []models.ItemContainer{}}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(emptyContainer, nil).Once()

		// Como não existe, ele deve criar um novo (AddItem) com quantidade 5
		iRepo.On("AddItem", mock.Anything, mock.MatchedBy(func(item *models.ItemContainer) bool {
			return item.Quantity == 5 && item.ProductID == 1
		})).Return(nil).Once()

		err := svc.AddProduct(ctx, 1, 1, 5)
		assert.NoError(t, err)
	})
}

func TestContainerService_UpdateProductQuantity(t *testing.T) {
	ctx := context.Background()

	t.Run("Corner Case - Product Not In Container", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)

		existingContainer := &models.Container{
			Items: []models.ItemContainer{{ProductID: 1, Quantity: 10}},
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(existingContainer, nil).Once()

		// Tenta atualizar o produto 99 (que não está no container)
		err := svc.UpdateProductQuantity(ctx, 1, 99, 50)
		assert.ErrorIs(t, err, service.ErrItemNotFound)
	})

	t.Run("Happy Path - Update Existent Item", func(t *testing.T) {
		svc, cRepo, iRepo, _ := setupService(t)

		existingContainer := &models.Container{
			Items: []models.ItemContainer{{ProductID: 99, Quantity: 10}},
		}
		cRepo.On("GetByID", mock.Anything, uint(1)).Return(existingContainer, nil).Once()

		iRepo.On("UpdateQuantity", mock.Anything, uint(1), uint(99), 50).Return(nil).Once()

		err := svc.UpdateProductQuantity(ctx, 1, 99, 50)
		assert.NoError(t, err)
	})
}

// --- STANDARD CRUD TESTS ---

func TestContainerService_GetByID(t *testing.T) {
	svc, cRepo, _, _ := setupService(t)
	ctx := context.Background()

	expectedContainer := &models.Container{Name: "Found Me"}
	cRepo.On("GetByID", mock.Anything, uint(1)).Return(expectedContainer, nil).Once()

	result, err := svc.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, "Found Me", result.Name)
}

func TestContainerService_GetAll(t *testing.T) {
	svc, cRepo, _, _ := setupService(t)
	ctx := context.Background()

	expectedList := []models.Container{{Name: "A"}, {Name: "B"}}
	cRepo.On("GetAll", mock.Anything, 1, 10).Return(expectedList, nil).Once()

	results, err := svc.GetAll(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestContainerService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)
		cRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		err := svc.Delete(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("Error Path", func(t *testing.T) {
		svc, cRepo, _, _ := setupService(t)
		cRepo.On("Delete", mock.Anything, uint(1)).Return(errors.New("db error")).Once()

		err := svc.Delete(ctx, 1)
		assert.Error(t, err)
	})
}

func TestContainerService_RemoveProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		svc, _, iRepo, _ := setupService(t)
		iRepo.On("RemoveItem", mock.Anything, uint(1), uint(99)).Return(nil).Once()

		err := svc.RemoveProduct(ctx, 1, 99)
		assert.NoError(t, err)
	})

	t.Run("Error Path", func(t *testing.T) {
		svc, _, iRepo, _ := setupService(t)
		iRepo.On("RemoveItem", mock.Anything, uint(1), uint(99)).Return(errors.New("delete failed")).Once()

		err := svc.RemoveProduct(ctx, 1, 99)
		assert.Error(t, err)
	})
}
