package importer

import (
	"context"
	"log/slog"

	"layoutmastergo/internal/service"

	"gorm.io/gorm"
)

// ContainerPayload representa o pacote de dados enviado aos workers
type ContainerPayload struct {
	ContainerName string
	Items         map[string]int // JAN Code -> Quantidade
}

// ExcelImporter encapsula as dependências para importar dados
type ExcelImporter struct {
	db           *gorm.DB
	productSvc   service.ProductService
	containerSvc service.ContainerService
	log          *slog.Logger
	ctx          context.Context
}

// NewExcelImporter cria uma nova instância do importador
func NewExcelImporter(db *gorm.DB, pSvc service.ProductService, cSvc service.ContainerService, log *slog.Logger) *ExcelImporter {
	return &ExcelImporter{
		db:           db,
		productSvc:   pSvc,
		containerSvc: cSvc,
		log:          log,
		ctx:          context.Background(),
	}
}

// Run é o método principal que orquestra as fases de importação
func (i *ExcelImporter) Run(filePath string) error {
	// 1. Lê as linhas do Excel
	rows, err := i.readExcelRows(filePath)
	if err != nil {
		return err
	}

	// 2. Fase 1: Sincroniza Produtos e gera cache em memória (Síncrono)
	janToIDCache := i.syncProducts(rows)

	// 3. Fase 2: Processa Contêineres e Itens (Assíncrono via Worker Pool)
	i.processContainers(rows, janToIDCache)

	return nil
}
