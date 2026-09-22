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

type ContainerController struct {
	svc service.ContainerService
	log *slog.Logger
}

type ContainerHandler interface {
	Create(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)

	AddProduct(c *gin.Context)
	UpdateProductQuantity(c *gin.Context)
	RemoveProduct(c *gin.Context)
}

type AddProductRequest struct {
	ProductID uint `json:"product_id" example:"1"`
	Quantity  int  `json:"quantity" example:"10"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" example:"15"`
}

func NewContainerController(svc service.ContainerService, log *slog.Logger) *ContainerController {
	return &ContainerController{
		svc: svc,
		log: log,
	}
}

func (ctrl *ContainerController) Create(c *gin.Context) {
	var req models.Container
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.log.Warn("invalid json for container creation", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos. Verifique se o Nome e a Data foram preenchidos corretamente."})
		return
	}

	if err := ctrl.svc.Create(c.Request.Context(), &req); err != nil {
		ctrl.log.Error("failed to create container", slog.String("error", err.Error()))

		if errors.Is(err, service.ErrContainerDuplicate) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrContainerBlankName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao criar contêiner."})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (ctrl *ContainerController) GetAll(c *gin.Context) {
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

	var containers []models.Container
	if search != "" {
		containers, err = ctrl.svc.Search(c.Request.Context(), search, page, pageSize)
	} else {
		containers, err = ctrl.svc.GetAll(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		ctrl.log.Error("failed to get containers", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao recuperar a lista de contêineres."})
		return
	}

	c.JSON(http.StatusOK, containers)
}

func (ctrl *ContainerController) GetByID(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	container, err := ctrl.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		ctrl.log.Warn("container not found", slog.Int("id", int(id)))
		c.JSON(http.StatusNotFound, gin.H{"error": "Contêiner não encontrado."})
		return
	}

	c.JSON(http.StatusOK, container)
}

func (ctrl *ContainerController) Update(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	var req models.Container
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos para atualização do contêiner."})
		return
	}

	req.ID = id

	if err := ctrl.svc.Update(c.Request.Context(), id, &req); err != nil {
		ctrl.log.Error("failed to update container", slog.Int("id", int(id)), slog.String("error", err.Error()))

		if errors.Is(err, service.ErrContainerDuplicate) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrContainerBlankName) || errors.Is(err, service.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao atualizar contêiner."})
		return
	}

	c.JSON(http.StatusOK, req)
}

func (ctrl *ContainerController) Delete(c *gin.Context) {
	id, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	if err := ctrl.svc.Delete(c.Request.Context(), id); err != nil {
		ctrl.log.Error("failed to delete container", slog.Int("id", int(id)), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao excluir contêiner."})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *ContainerController) AddProduct(c *gin.Context) {
	containerID, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	var req AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos para adição de produto."})
		return
	}

	if err := ctrl.svc.AddProduct(c.Request.Context(), containerID, req.ProductID, req.Quantity); err != nil {
		ctrl.log.Error("failed to add product to container", slog.Int("container_id", int(containerID)), slog.String("error", err.Error()))

		if errors.Is(err, service.ErrInvalidQuantity) || errors.Is(err, service.ErrProductNotFound) || errors.Is(err, service.ErrContainerNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao adicionar produto ao contêiner."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Produto adicionado ao contêiner com sucesso"})
}

func (ctrl *ContainerController) UpdateProductQuantity(c *gin.Context) {
	containerID, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	productID, err := ctrl.parseID(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de produto inválido."})
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantidade informada inválida."})
		return
	}

	if err := ctrl.svc.UpdateProductQuantity(c.Request.Context(), containerID, productID, req.Quantity); err != nil {
		ctrl.log.Error("failed to update product quantity", slog.Int("container_id", int(containerID)), slog.String("error", err.Error()))

		if errors.Is(err, service.ErrInvalidQuantity) || errors.Is(err, service.ErrItemNotFound) || errors.Is(err, service.ErrContainerNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao atualizar quantidade do produto."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quantidade atualizada com sucesso"})
}

func (ctrl *ContainerController) RemoveProduct(c *gin.Context) {
	containerID, err := ctrl.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de contêiner inválido."})
		return
	}

	productID, err := ctrl.parseID(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de produto inválido."})
		return
	}

	if err := ctrl.svc.RemoveProduct(c.Request.Context(), containerID, productID); err != nil {
		ctrl.log.Error("failed to remove product from container", slog.Int("container_id", int(containerID)), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao remover produto do contêiner."})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *ContainerController) parseID(idParam string) (uint, error) {
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctrl.log.Warn("failed to parse ID", slog.String("param", idParam))
		return 0, err
	}
	return uint(id), nil
}
