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

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) Create(ctx context.Context, p *models.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockProductService) GetAll(ctx context.Context, page, pageSize int) ([]models.Product, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) Search(ctx context.Context, query string, page, pageSize int) ([]models.Product, error) {
	args := m.Called(ctx, query, page, pageSize)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductService) Update(ctx context.Context, id uint, p *models.Product) error {
	args := m.Called(ctx, id, p)
	return args.Error(0)
}

func (m *MockProductService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupProductTestRouter(mockSvc *MockProductService) (*gin.Engine, *controller.ProductController) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctrl := controller.NewProductController(mockSvc, log)
	return router, ctrl
}

func TestProductController_GetAll(t *testing.T) {
	t.Run("Happy Path - Default Pagination", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Product{{}}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Happy Path - With Search Parameter", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products", ctrl.GetAll)

		mockSvc.On("Search", mock.Anything, "Mouse", 1, 10).Return([]models.Product{{Name: "Mouse"}}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products?search=Mouse", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Corner Case - Invalid Query Params fallback", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Product{}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products?page=abc&pageSize=-5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error on GetAll", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products", ctrl.GetAll)

		mockSvc.On("GetAll", mock.Anything, 1, 10).Return([]models.Product{}, errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error on Search", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products", ctrl.GetAll)

		mockSvc.On("Search", mock.Anything, "Keyboard", 1, 10).Return([]models.Product{}, errors.New("db error"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products?search=Keyboard", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestProductController_GetByID(t *testing.T) {
	t.Run("Happy Path - Valid ID", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products/:id", ctrl.GetByID)

		mockSvc.On("GetByID", mock.Anything, uint(1)).Return(&models.Product{}, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid ID Format", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products/:id", ctrl.GetByID)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Not Found", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.GET("/api/v1/products/:id", ctrl.GetByID)

		mockSvc.On("GetByID", mock.Anything, uint(99)).Return((*models.Product)(nil), errors.New("not found"))

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/99", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestProductController_Create(t *testing.T) {
	t.Run("Happy Path - Valid Creation", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.POST("/api/v1/products", ctrl.Create)

		mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*models.Product")).Return(nil)

		body := []byte(`{"name": "Test Product", "jan_code": "12345"}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid JSON", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.POST("/api/v1/products", ctrl.Create)

		body := []byte(`{bad-json}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.POST("/api/v1/products", ctrl.Create)

		mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*models.Product")).Return(errors.New("db error"))

		body := []byte(`{"name": "Test Product", "jan_code": "12345"}`)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestProductController_Update(t *testing.T) {
	t.Run("Happy Path - Successful Update", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.PUT("/api/v1/products/:id", ctrl.Update)

		mockSvc.On("Update", mock.Anything, uint(1), mock.AnythingOfType("*models.Product")).Return(nil)

		body := []byte(`{"name": "Updated Product", "jan_code": "12345"}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid ID Format", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.PUT("/api/v1/products/:id", ctrl.Update)

		req, _ := http.NewRequest(http.MethodPut, "/api/v1/products/abc", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid JSON", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.PUT("/api/v1/products/:id", ctrl.Update)

		req, _ := http.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer([]byte(`{bad-json}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.PUT("/api/v1/products/:id", ctrl.Update)

		mockSvc.On("Update", mock.Anything, uint(1), mock.AnythingOfType("*models.Product")).Return(errors.New("db error"))

		body := []byte(`{"name": "Updated Product", "jan_code": "12345"}`)
		req, _ := http.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestProductController_Delete(t *testing.T) {
	t.Run("Happy Path - Successful Deletion", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.DELETE("/api/v1/products/:id", ctrl.Delete)

		mockSvc.On("Delete", mock.Anything, uint(1)).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Invalid ID Format", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.DELETE("/api/v1/products/:id", ctrl.Delete)

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("Bad Path - Service Error", func(t *testing.T) {
		mockSvc := new(MockProductService)
		router, ctrl := setupProductTestRouter(mockSvc)
		router.DELETE("/api/v1/products/:id", ctrl.Delete)

		mockSvc.On("Delete", mock.Anything, uint(1)).Return(errors.New("db error"))

		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
