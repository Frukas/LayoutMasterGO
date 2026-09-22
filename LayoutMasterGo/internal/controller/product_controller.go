package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"layoutmastergo/internal/models"
	"layoutmastergo/internal/service"
)

type ProductHandler interface {
	Create(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type ProductController struct {
	svc service.ProductService
	log *slog.Logger
}

func NewProductController(svc service.ProductService, log *slog.Logger) *ProductController {
	return &ProductController{
		svc: svc,
		log: log,
	}
}

func (ctrl *ProductController) Create(c *gin.Context) {
	var req models.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.log.Warn("invalid json for product creation", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos. Verifique se o Nome e o JanCode foram preenchidos corretamente."})
		return
	}

	if err := ctrl.svc.Create(c.Request.Context(), &req); err != nil {
		ctrl.log.Error("failed to create product", slog.String("error", err.Error()))

		if errors.Is(err, service.ErrProductDuplicate) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidProductName) || errors.Is(err, service.ErrInvalidJanCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao criar produto."})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (ctrl *ProductController) GetAll(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var products []models.Product
	if search != "" {
		products, err = ctrl.svc.Search(c.Request.Context(), search, page, pageSize)
	} else {
		products, err = ctrl.svc.GetAll(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		ctrl.log.Error("failed to get products", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao recuperar a lista de produtos."})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (ctrl *ProductController) GetByID(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de produto inválido."})
		return
	}

	product, err := ctrl.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		ctrl.log.Warn("product not found", slog.Int("id", int(id)))
		c.JSON(http.StatusNotFound, gin.H{"error": "Produto não encontrado."})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (ctrl *ProductController) Update(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de produto inválido."})
		return
	}

	var req models.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos para atualização."})
		return
	}

	if err := ctrl.svc.Update(c.Request.Context(), id, &req); err != nil {
		ctrl.log.Error("failed to update product", slog.Int("id", int(id)), slog.String("error", err.Error()))

		if errors.Is(err, service.ErrProductDuplicate) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrInvalidProductName) || errors.Is(err, service.ErrInvalidJanCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao atualizar produto."})
		return
	}

	c.JSON(http.StatusOK, req)
}

func (ctrl *ProductController) Delete(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de produto inválido."})
		return
	}

	if err := ctrl.svc.Delete(c.Request.Context(), id); err != nil {
		ctrl.log.Error("failed to delete product", slog.Int("id", int(id)), slog.String("error", err.Error()))

		if errors.Is(err, service.ErrProductInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao excluir produto."})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *ProductController) parseID(idParam string) (uint, error) {
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctrl.log.Warn("failed to parse ID", slog.String("param", idParam))
		return 0, err
	}
	return uint(id), nil
}
