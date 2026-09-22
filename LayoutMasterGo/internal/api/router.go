package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "layoutmastergo/docs"
	"layoutmastergo/internal/api/middleware"
	"layoutmastergo/internal/controller"
	"layoutmastergo/web"
)

type AppControllers struct {
	Container controller.ContainerHandler
	Product   controller.ProductHandler
}

func SetupRoutes(router *gin.Engine, controllers AppControllers) {
	router.Use(middleware.CORSMiddleware())

	// Serve os arquivos estáticos (JS e HTML templates) a partir da pasta embutida web/
	router.StaticFS("/static", http.FS(web.Files))

	// Rota raiz entrega o Shell principal da SPA
	router.GET("/", func(c *gin.Context) {
		file, err := web.Files.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Erro ao carregar index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", file)
	})

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API REST V1 (Retornos puramente JSON)
	v1 := router.Group("/api/v1")
	{
		containers := v1.Group("/containers")
		{
			containers.POST("", controllers.Container.Create)
			containers.GET("", controllers.Container.GetAll)
			containers.GET("/:id", controllers.Container.GetByID)
			containers.PUT("/:id", controllers.Container.Update)
			containers.DELETE("/:id", controllers.Container.Delete)

			containers.POST("/:id/items", controllers.Container.AddProduct)
			containers.PUT("/:id/items/:product_id", controllers.Container.UpdateProductQuantity)
			containers.DELETE("/:id/items/:product_id", controllers.Container.RemoveProduct)
		}

		products := v1.Group("/products")
		{
			products.POST("", controllers.Product.Create)
			products.GET("", controllers.Product.GetAll)
			products.GET("/:id", controllers.Product.GetByID)
			products.PUT("/:id", controllers.Product.Update)
			products.DELETE("/:id", controllers.Product.Delete)
		}
	}
}
