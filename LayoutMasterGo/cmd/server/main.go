package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"

	"layoutmastergo/internal/api"
	"layoutmastergo/internal/controller"
	"layoutmastergo/internal/database"
	"layoutmastergo/internal/models"
	"layoutmastergo/internal/repository"
	"layoutmastergo/internal/service"
)

// @title           Layout Master Go API
// @version         1.0
// @description     API RESTful para gerenciamento de contêineres e produtos desenvolvida com Gin, GORM e Clean Architecture.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Suporte técnico
// @contact.email   suporte@layoutmaster.com

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Digite "Bearer " seguido do seu token JWT
func main() {

	if err := godotenv.Load(); err != nil {
		slog.Info("Arquivo .env não encontrado, utilizando padrões do sistema")
	}
	// 1. Configuração de Logs
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Instanciação do Dialector (SQLite)
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}
	dialector := database.NewSQLiteDialector(dbPath)

	// 3. Inicialização agnóstica do Banco de Dados
	db, err := database.NewConnection(dialector)
	if err != nil {
		log.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Execução de Migrações
	err = db.AutoMigrate(
		&models.Product{},
		&models.Container{},
		&models.ItemContainer{},
	)
	if err != nil {
		log.Error("Failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("Database connected and migrated successfully")

	// 5. Injeção de Dependências (Camada de Repositório)
	cRepo := repository.NewContainerRepository(db, log)
	iRepo := repository.NewItemContainerRepository(db, log)
	pRepo := repository.NewProductRepository(db, log)

	// 6. Injeção de Dependências (Camada de Serviço)
	containerSvc := service.NewContainerService(cRepo, iRepo, pRepo, log)
	productSvc := service.NewProductService(pRepo, log)

	// 7. Injeção de Dependências (Camada de Controller)
	containerCtrl := controller.NewContainerController(containerSvc, log)
	productCtrl := controller.NewProductController(productSvc, log)

	// 8. Configuração do Framework Web (Gin)
	router := gin.Default()

	// 9. Registro Centralizado de Rotas
	controllers := api.AppControllers{
		Container: containerCtrl,
		Product:   productCtrl,
	}
	api.SetupRoutes(router, controllers)

	// 10. Inicialização do Servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("Starting API server with Gin", slog.String("port", port))

	if err := router.Run(":" + port); err != nil {
		log.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
