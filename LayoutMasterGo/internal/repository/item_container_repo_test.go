package repository_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"layoutmastergo/internal/models"
	"layoutmastergo/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBItemContainer(t *testing.T) (*gorm.DB, repository.ItemContainerRepository) {
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
	return db, repository.NewItemContainerRepository(db, silentLogger)
}

func helperDependencies(db *gorm.DB) (models.Container, models.Product) {
	c := models.Container{Name: "Warehouse", Data: time.Now()}
	p := models.Product{Name: "Keyboard", JanCode: "123"}
	db.Create(&c)
	db.Create(&p)
	return c, p
}

func TestItemContainerRepository_AddItem(t *testing.T) {
	db, repo := setupTestDBItemContainer(t)
	ctx := context.Background()

	c, p := helperDependencies(db)
	newItem := &models.ItemContainer{ContainerID: c.ID, ProductID: p.ID, Quantity: 100}

	// 1. Happy Path
	err := repo.AddItem(ctx, newItem)
	assert.NoError(t, err)

	var saved models.ItemContainer
	db.Where("container_id = ? AND product_id = ?", c.ID, p.ID).First(&saved)
	assert.Equal(t, 100, saved.Quantity)

	// 2. Duplicate PK Error
	err = repo.AddItem(ctx, newItem)
	assert.Error(t, err, "expected duplicate key error")

	// 3. Foreign Key Error
	invalidItem := &models.ItemContainer{ContainerID: 9999, ProductID: p.ID, Quantity: 50}
	err = repo.AddItem(ctx, invalidItem)
	assert.Error(t, err, "expected FK constraint error")
}

func TestItemContainerRepository_UpdateQuantity(t *testing.T) {
	db, repo := setupTestDBItemContainer(t)
	ctx := context.Background()

	c, p := helperDependencies(db)
	db.Create(&models.ItemContainer{ContainerID: c.ID, ProductID: p.ID, Quantity: 50})

	err := repo.UpdateQuantity(ctx, c.ID, p.ID, 300)
	assert.NoError(t, err)

	var updated models.ItemContainer
	db.Where("container_id = ? AND product_id = ?", c.ID, p.ID).First(&updated)
	assert.Equal(t, 300, updated.Quantity)
}

func TestItemContainerRepository_RemoveItem(t *testing.T) {
	db, repo := setupTestDBItemContainer(t)
	ctx := context.Background()

	c, p1 := helperDependencies(db)
	p2 := models.Product{Name: "Mouse", JanCode: "54321"}
	db.Create(&p2)

	db.Create(&models.ItemContainer{ContainerID: c.ID, ProductID: p1.ID, Quantity: 10})
	db.Create(&models.ItemContainer{ContainerID: c.ID, ProductID: p2.ID, Quantity: 20})

	err := repo.RemoveItem(ctx, c.ID, p1.ID)
	assert.NoError(t, err)

	// Verifies p1 was removed
	err = db.Where("container_id = ? AND product_id = ?", c.ID, p1.ID).First(&models.ItemContainer{}).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Verifies p2 remains untouched
	var check models.ItemContainer
	err = db.Where("container_id = ? AND product_id = ?", c.ID, p2.ID).First(&check).Error
	assert.NoError(t, err)
	assert.Equal(t, 20, check.Quantity)
}
