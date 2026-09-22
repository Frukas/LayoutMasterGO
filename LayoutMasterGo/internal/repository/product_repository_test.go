package repository_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"

	"layoutmastergo/internal/models"
	"layoutmastergo/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBProduct(t *testing.T) (*gorm.DB, repository.ProductRepository) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	db.Exec("PRAGMA foreign_keys = ON")
	db.AutoMigrate(&models.Product{}, &models.Container{}, &models.ItemContainer{})

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return db, repository.NewProductRepository(db, silentLogger)
}

func TestProductRepository_Save(t *testing.T) {
	_, repo := setupTestDBProduct(t)
	ctx := context.Background()

	newProduct := &models.Product{Name: "Keyboard", JanCode: "123"}
	err := repo.Save(ctx, newProduct)

	assert.NoError(t, err)
	assert.NotZero(t, newProduct.ID)
}

func TestProductRepository_Save_DuplicateError(t *testing.T) {
	db, repo := setupTestDBProduct(t)
	ctx := context.Background()

	db.Create(&models.Product{Name: "Unique", JanCode: "001"})

	errName := repo.Save(ctx, &models.Product{Name: "Unique", JanCode: "002"})
	assert.Error(t, errName, "expected UNIQUE constraint error on Name")

	errJan := repo.Save(ctx, &models.Product{Name: "Another", JanCode: "001"})
	assert.Error(t, errJan, "expected UNIQUE constraint error on JanCode")
}

func TestProductRepository_GetByID(t *testing.T) {
	db, repo := setupTestDBProduct(t)
	ctx := context.Background()

	seeded := models.Product{Name: "Mouse", JanCode: "111"}
	db.Create(&seeded)

	result, err := repo.GetByID(ctx, seeded.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Mouse", result.Name)
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	_, repo := setupTestDBProduct(t)
	_, err := repo.GetByID(context.Background(), 9999)
	assert.Error(t, err)
}

func TestProductRepository_GetAll_And_PaginationEdgeCases(t *testing.T) {
	db, repo := setupTestDBProduct(t)
	ctx := context.Background()

	db.Create(&models.Product{Name: "A", JanCode: "1"})
	db.Create(&models.Product{Name: "B", JanCode: "2"})
	db.Create(&models.Product{Name: "C", JanCode: "3"})

	results, err := repo.GetAll(ctx, 1, 2)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	resultsEdge, err := repo.GetAll(ctx, -1, 500)
	assert.NoError(t, err)
	assert.Len(t, resultsEdge, 3)
}

func TestProductRepository_Update(t *testing.T) {
	db, repo := setupTestDBProduct(t)
	ctx := context.Background()

	seeded := models.Product{Name: "Old", JanCode: "000"}
	db.Create(&seeded)

	err := repo.Update(ctx, seeded.ID, &models.Product{Name: "New"})
	assert.NoError(t, err)

	var updated models.Product
	db.First(&updated, seeded.ID)
	assert.Equal(t, "New", updated.Name)
}

func TestProductRepository_Delete(t *testing.T) {
	db, repo := setupTestDBProduct(t)
	ctx := context.Background()

	seeded := models.Product{Name: "Del", JanCode: "999"}
	db.Create(&seeded)

	err := repo.Delete(ctx, seeded.ID)
	assert.NoError(t, err)

	err = db.First(&models.Product{}, seeded.ID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
