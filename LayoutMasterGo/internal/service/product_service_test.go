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

func setupProductService(t *testing.T) (service.ProductService, *mocks.ProductRepository) {
	pRepo := mocks.NewProductRepository(t)
	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewProductService(pRepo, silentLogger)
	return svc, pRepo
}

func TestProductService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Negative Path - Empty Name", func(t *testing.T) {
		svc, _ := setupProductService(t)
		product := &models.Product{Name: "   ", JanCode: "12345"}

		err := svc.Create(ctx, product)
		assert.ErrorIs(t, err, service.ErrInvalidProductName)
	})

	t.Run("Negative Path - Empty JanCode", func(t *testing.T) {
		svc, _ := setupProductService(t)
		product := &models.Product{Name: "Keyboard", JanCode: ""}

		err := svc.Create(ctx, product)
		assert.ErrorIs(t, err, service.ErrInvalidJanCode)
	})

	t.Run("Happy Path - Sanitization and Save", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		product := &models.Product{Name: "  Gaming Mouse  ", JanCode: "  JAN-999  "}

		pRepo.On("Save", mock.Anything, mock.MatchedBy(func(p *models.Product) bool {
			return p.Name == "Gaming Mouse" && p.JanCode == "JAN-999"
		})).Return(nil).Once()

		err := svc.Create(ctx, product)
		assert.NoError(t, err)
	})

	t.Run("Error Path - Repository Failure", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		product := &models.Product{Name: "Valid Name", JanCode: "123"}
		expectedErr := errors.New("db error")

		pRepo.On("Save", mock.Anything, product).Return(expectedErr).Once()

		err := svc.Create(ctx, product)
		assert.ErrorIs(t, err, expectedErr)
	})

	// NOVO TESTE: Valida se o banco disparar erro de duplicidade, o service devolve o Sentinela
	t.Run("Error Path - Duplicate Product", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		product := &models.Product{Name: "Valid Name", JanCode: "123"}

		pRepo.On("Save", mock.Anything, product).Return(errors.New("UNIQUE constraint failed")).Once()

		err := svc.Create(ctx, product)
		assert.ErrorIs(t, err, service.ErrProductDuplicate)
	})
}

func TestProductService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Negative Path - Update with empty Name", func(t *testing.T) {
		svc, _ := setupProductService(t)
		updateData := &models.Product{Name: "   "}

		err := svc.Update(ctx, 1, updateData)
		assert.ErrorIs(t, err, service.ErrInvalidProductName)
	})

	t.Run("Negative Path - Update with empty JanCode", func(t *testing.T) {
		svc, _ := setupProductService(t)
		updateData := &models.Product{JanCode: "   "}

		err := svc.Update(ctx, 1, updateData)
		assert.ErrorIs(t, err, service.ErrInvalidJanCode)
	})

	t.Run("Happy Path - Partial Update with Sanitization", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		updateData := &models.Product{Name: " New Name "}

		pRepo.On("Update", mock.Anything, uint(1), mock.MatchedBy(func(p *models.Product) bool {
			return p.Name == "New Name"
		})).Return(nil).Once()

		err := svc.Update(ctx, 1, updateData)
		assert.NoError(t, err)
	})

	t.Run("Error Path - Repository Failure", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		updateData := &models.Product{Name: "Valid Name"}
		expectedErr := errors.New("db error")

		pRepo.On("Update", mock.Anything, uint(1), updateData).Return(expectedErr).Once()

		err := svc.Update(ctx, 1, updateData)
		assert.ErrorIs(t, err, expectedErr)
	})

	// NOVO TESTE: Valida erro de duplicidade no Update
	t.Run("Error Path - Duplicate Product on Update", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		updateData := &models.Product{Name: "Duplicate Name"}

		pRepo.On("Update", mock.Anything, uint(1), updateData).Return(errors.New("duplicate entry")).Once()

		err := svc.Update(ctx, 1, updateData)
		assert.ErrorIs(t, err, service.ErrProductDuplicate)
	})
}

func TestProductService_GetByID(t *testing.T) {
	svc, pRepo := setupProductService(t)
	ctx := context.Background()

	expectedProduct := &models.Product{Name: "Found Product"}
	pRepo.On("GetByID", mock.Anything, uint(1)).Return(expectedProduct, nil).Once()

	result, err := svc.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, "Found Product", result.Name)
}

func TestProductService_GetAll(t *testing.T) {
	svc, pRepo := setupProductService(t)
	ctx := context.Background()

	expectedList := []models.Product{{Name: "P1"}, {Name: "P2"}}
	pRepo.On("GetAll", mock.Anything, 1, 10).Return(expectedList, nil).Once()

	results, err := svc.GetAll(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestProductService_Search(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path - Valid Query Delegates To Search", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		expectedList := []models.Product{{Name: "Gaming Mouse"}}

		pRepo.On("Search", mock.Anything, "Mouse", 1, 10).Return(expectedList, nil).Once()

		results, err := svc.Search(ctx, "  Mouse  ", 1, 10)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Gaming Mouse", results[0].Name)
	})

	t.Run("Fallback Path - Empty Query Delegates To GetAll", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		expectedList := []models.Product{{Name: "P1"}, {Name: "P2"}}

		pRepo.On("GetAll", mock.Anything, 1, 10).Return(expectedList, nil).Once()

		results, err := svc.Search(ctx, "   ", 1, 10)

		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("Error Path - Repository Search Failure", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		expectedErr := errors.New("search error")

		pRepo.On("Search", mock.Anything, "Keyboard", 1, 10).Return([]models.Product{}, expectedErr).Once()

		_, err := svc.Search(ctx, "Keyboard", 1, 10)

		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestProductService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		pRepo.On("Delete", mock.Anything, uint(1)).Return(nil).Once()

		err := svc.Delete(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("Error Path", func(t *testing.T) {
		svc, pRepo := setupProductService(t)
		expectedErr := errors.New("db error")

		pRepo.On("Delete", mock.Anything, uint(1)).Return(expectedErr).Once()

		err := svc.Delete(ctx, 1)
		assert.ErrorIs(t, err, expectedErr)
	})

	// NOVO TESTE: Produto não pode ser apagado se estiver em um contêiner
	t.Run("Error Path - Product In Use", func(t *testing.T) {
		svc, pRepo := setupProductService(t)

		pRepo.On("Delete", mock.Anything, uint(1)).Return(errors.New("foreign key constraint failed")).Once()

		err := svc.Delete(ctx, 1)
		assert.ErrorIs(t, err, service.ErrProductInUse)
	})
}
