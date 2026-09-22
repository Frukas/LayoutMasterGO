package models

import (
	"time"

	"gorm.io/gorm"
)

// Product representa um produto cadastrado no sistema
type Product struct {
	gorm.Model
	Name    string          `gorm:"size:255;not null;unique" json:"name" binding:"required" example:"Detergente Líquido"`
	JanCode string          `gorm:"size:255;not null;unique" json:"jan_code" binding:"required" example:"4901234567890"`
	Items   []ItemContainer `gorm:"foreignKey:ProductID" json:"items,omitempty"`
}

// Container representa um armazenamento ou lote físico
type Container struct {
	gorm.Model
	Name   string          `gorm:"size:255;not null;unique" json:"name" binding:"required" example:"Lote A-100"`
	Data   time.Time       `gorm:"not null" json:"data" binding:"required" example:"2026-09-09T10:00:00Z"`
	Status string          `json:"status" example:"Em trânsito"`
	Items  []ItemContainer `gorm:"foreignKey:ContainerID" json:"items,omitempty"`
}

// ItemContainer é a tabela pivô que mapeia os produtos aos contêineres com suas quantidades
type ItemContainer struct {
	ProductID   uint       `gorm:"primaryKey" json:"product_id" example:"1"`
	ContainerID uint       `gorm:"primaryKey" json:"container_id" example:"10"`
	Quantity    int        `gorm:"not null" json:"quantity" binding:"required,gt=0" example:"50"`
	Product     *Product   `gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"product,omitempty"`
	Container   *Container `gorm:"foreignKey:ContainerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"container,omitempty"`

	CreatedAt time.Time      `json:"created_at" swaggerignore:"true"`
	UpdatedAt time.Time      `json:"updated_at" swaggerignore:"true"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at" swaggerignore:"true"`
}
