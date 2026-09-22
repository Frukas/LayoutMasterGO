package controller_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"layoutmastergo/internal/controller"
	"layoutmastergo/internal/models"
)

type MockContainerService struct {
	mock.Mock
}

func (m *MockContainerService) Create(ctx context.Context, c *models.Container) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockContainerService) GetAll(ctx context.Context, page, pageSize int) ([]models.Container, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]models.Container), args.Error(1)
}

func (m *MockContainerService) Search(ctx context.Context, query string, page, pageSize int) ([]models.Container, error) {
	args := m.Called(ctx, query, page, pageSize)
	return args.Get(0).([]models.Container), args.Error(1)
}

func (m *MockContainerService) GetByID(ctx context.Context, id uint) (*models.Container, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Container), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContainerService) Update(ctx context.Context, id uint, c *models.Container) error {
	args := m.Called(ctx, id, c)
	return args.Error(0)
}

func (m *MockContainerService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerService) AddProduct(ctx context.Context, containerID, productID uint, quantity int) error {
	args := m.Called(ctx, containerID, productID, quantity)
	return args.Error(0)
}

func (m *MockContainerService) UpdateProductQuantity(ctx context.Context, containerID, productID uint, quantity int) error {
	args := m.Called(ctx, containerID, productID, quantity)
	return args.Error(0)
}

func (m *MockContainerService) RemoveProduct(ctx context.Context, containerID, productID uint) error {
	args := m.Called(ctx, containerID, productID)
	return args.Error(0)
}

func setupTestRouter(mockSvc *MockContainerService) (*gin.Engine, *controller.ContainerController) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctrl := controller.NewContainerController(mockSvc, log)
	return router, ctrl
}

func TestContainerController_GetAll(t *testing.T) {
	t.Run("Happy Path - Default Pagination", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Container{{}}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Happy Path - With Search Parameter", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers", ctrl.GetAll)

		mockSvc.On("Search", mock.Anything, "Dry", 1, 10).Return([]models.Container{{Name: "Dry Container"}}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers?search=Dry", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Corner Case - Invalid Query Params fallback to default", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Container{}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers?page=abc&pageSize=-5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error on GetAll", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Container{}, errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error on Search", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers", ctrl.GetAll)

		mockSvc.On("Search", mock.Anything, "Dry", 1, 10).Return([]models.Container{}, errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers?search=Dry", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_GetByID(t *testing.T) {
	t.Run("Happy Path - Valid ID", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers/:id", ctrl.GetByID)

		mockSvc.On("GetByID", mock.Anything, uint(1)).Return(&models.Container{}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid ID Format", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers/:id", ctrl.GetByID)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Not Found", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.GET("/api/v1/containers/:id", ctrl.GetByID)

		mockSvc.On("GetByID", mock.Anything, uint(99)).Return((*models.Container)(nil), errors.New("not found"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/containers/99", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_Create(t *testing.T) {
	t.Run("Happy Path - Valid Creation", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers", ctrl.Create)

		mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*models.Container")).Return(nil)

		body := []byte(`{"name": "Container 01", "data": "2026-09-10T10:00:00Z", "status": "Pending"}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid JSON", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers", ctrl.Create)

		body := []byte(`{bad-json}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_Update(t *testing.T) {
	t.Run("Happy Path - Successful Update", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id", ctrl.Update)

		mockSvc.On("Update", mock.Anything, uint(1), mock.AnythingOfType("*models.Container")).Return(nil)

		body := []byte(`{"name": "Updated Container", "data": "2026-09-10T10:00:00Z", "status": "Active"}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid JSON Payload", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id", ctrl.Update)

		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1", bytes.NewBuffer([]byte(`{bad-json}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id", ctrl.Update)

		mockSvc.On("Update", mock.Anything, uint(1), mock.AnythingOfType("*models.Container")).Return(errors.New("db error"))

		body := []byte(`{"name": "Container 01", "data": "2026-09-10T10:00:00Z", "status": "Active"}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_Delete(t *testing.T) {
	t.Run("Happy Path - Successful Deletion", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.DELETE("/api/v1/containers/:id", ctrl.Delete)

		mockSvc.On("Delete", mock.Anything, uint(1)).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/containers/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.DELETE("/api/v1/containers/:id", ctrl.Delete)

		mockSvc.On("Delete", mock.Anything, uint(1)).Return(errors.New("db error"))

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/containers/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContainerController_AddProduct(t *testing.T) {
	t.Run("Happy Path - Product Added Successfully", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers/:id/items", ctrl.AddProduct)

		mockSvc.On("AddProduct", mock.Anything, uint(1), uint(10), 5).Return(nil)

		body := []byte(`{"product_id": 10, "quantity": 5}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers/1/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid Container ID", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers/:id/items", ctrl.AddProduct)

		body := []byte(`{"product_id": 10, "quantity": 5}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers/abc/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Path - Invalid JSON Payload", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers/:id/items", ctrl.AddProduct)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers/1/items", bytes.NewBuffer([]byte(`{bad-json}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.POST("/api/v1/containers/:id/items", ctrl.AddProduct)

		mockSvc.On("AddProduct", mock.Anything, uint(1), uint(10), 5).Return(errors.New("db error"))

		body := []byte(`{"product_id": 10, "quantity": 5}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/containers/1/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_UpdateProductQuantity(t *testing.T) {
	t.Run("Happy Path - Quantity Updated Successfully", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id/items/:product_id", ctrl.UpdateProductQuantity)

		mockSvc.On("UpdateProductQuantity", mock.Anything, uint(1), uint(10), 20).Return(nil)

		body := []byte(`{"quantity": 20}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1/items/10", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid Product ID", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id/items/:product_id", ctrl.UpdateProductQuantity)

		body := []byte(`{"quantity": 20}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1/items/abc", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.PUT("/api/v1/containers/:id/items/:product_id", ctrl.UpdateProductQuantity)

		mockSvc.On("UpdateProductQuantity", mock.Anything, uint(1), uint(10), 20).Return(errors.New("db error"))

		body := []byte(`{"quantity": 20}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/containers/1/items/10", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestContainerController_RemoveProduct(t *testing.T) {
	t.Run("Happy Path - Product Removed Successfully", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.DELETE("/api/v1/containers/:id/items/:product_id", ctrl.RemoveProduct)

		mockSvc.On("RemoveProduct", mock.Anything, uint(1), uint(10)).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/containers/1/items/10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid Container ID", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.DELETE("/api/v1/containers/:id/items/:product_id", ctrl.RemoveProduct)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/containers/invalid/items/10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockContainerService)
		router, ctrl := setupTestRouter(mockSvc)
		router.DELETE("/api/v1/containers/:id/items/:product_id", ctrl.RemoveProduct)

		mockSvc.On("RemoveProduct", mock.Anything, uint(1), uint(10)).Return(errors.New("db error"))

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/containers/1/items/10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
