package main

import (
	"log/slog"
	"os"
	"time"

	"layoutmastergo/internal/importer"
	"layoutmastergo/internal/models"
	"layoutmastergo/internal/repository"
	"layoutmastergo/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// TEMPORIZADOR GERAL INICIA AQUI
	start := time.Now()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db := setupDatabase(log)

	pRepo := repository.NewProductRepository(db, log)
	cRepo := repository.NewContainerRepository(db, log)
	iRepo := repository.NewItemContainerRepository(db, log)

	productSvc := service.NewProductService(pRepo, log)
	containerSvc := service.NewContainerService(cRepo, iRepo, pRepo, log)

	excelImporter := importer.NewExcelImporter(db, productSvc, containerSvc, log)

	fileName := "data.xlsx"
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	}

	log.Info("Starting import process", slog.String("file", fileName))

	if err := excelImporter.Run(fileName); err != nil {
		log.Error("import process failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// TEMPORIZADOR GERAL FINALIZA AQUI
	log.Info("========================================")
	log.Info("🚀 IMPORT SUCCESSFULLY COMPLETED!")
	log.Info("⏱️ TOTAL DURATION", slog.String("total_time", time.Since(start).String()))
	log.Info("========================================")
}

func setupDatabase(log *slog.Logger) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")

	_ = db.AutoMigrate(&models.Product{}, &models.Container{}, &models.ItemContainer{})

	return db
}
