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

func setupTestDB(t *testing.T) (*gorm.DB, repository.ContainerRepository) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "failed to connect to test db")

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	db.Exec("PRAGMA foreign_keys = ON")
	db.AutoMigrate(&models.Container{}, &models.Product{}, &models.ItemContainer{})

	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return db, repository.NewContainerRepository(db, silentLogger)
}

func TestContainerRepository_Save(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	container := &models.Container{Name: "Warehouse A", Status: "Active"}
	err := repo.Save(ctx, container)

	assert.NoError(t, err)
	assert.NotZero(t, container.ID, "expected container ID to be generated")
}

func TestContainerRepository_Save_NegativePath_DuplicateName(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	db.Create(&models.Container{Name: "Unique Box"})

	duplicate := &models.Container{Name: "Unique Box"}
	err := repo.Save(ctx, duplicate)

	assert.Error(t, err, "expected UNIQUE constraint error")
}

func TestContainerRepository_GetByID(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	seededContainer := models.Container{Name: "Warehouse A"}
	db.Create(&seededContainer)

	result, err := repo.GetByID(ctx, seededContainer.ID)

	assert.NoError(t, err)
	assert.Equal(t, "Warehouse A", result.Name)
}

func TestContainerRepository_GetByID_NotFound(t *testing.T) {
	_, repo := setupTestDB(t)

	_, err := repo.GetByID(context.Background(), 9999)
	assert.Error(t, err)
}

func TestContainerRepository_GetAll_PaginationEdgeCases(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	db.Create(&models.Container{Name: "A"})
	db.Create(&models.Container{Name: "B"})
	db.Create(&models.Container{Name: "C"})

	results, err := repo.GetAll(ctx, -5, 9999) // Fallback test
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	resultsZero, err := repo.GetAll(ctx, 1, 0)
	assert.NoError(t, err)
	assert.Len(t, resultsZero, 3)
}

func TestContainerRepository_Update_BlackHoleCase(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	original := models.Container{Name: "Original"}
	db.Create(&original)

	emptyUpdate := &models.Container{}
	err := repo.Update(ctx, original.ID, emptyUpdate)
	assert.NoError(t, err)

	var check models.Container
	db.First(&check, original.ID)
	assert.Equal(t, "Original", check.Name, "empty update should not override data")
}

func TestContainerRepository_Delete(t *testing.T) {
	db, repo := setupTestDB(t)
	ctx := context.Background()

	seeded := models.Container{Name: "Delete Me"}
	db.Create(&seeded)

	err := repo.Delete(ctx, seeded.ID)
	assert.NoError(t, err)

	var check models.Container
	err = db.First(&check, seeded.ID).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "record should be soft deleted")
}
