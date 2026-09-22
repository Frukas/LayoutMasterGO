package importer

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"layoutmastergo/internal/models"

	"github.com/xuri/excelize/v2"
)

// readExcelRows abre o arquivo e valida a estrutura mínima na primeira aba encontrada
func (i *ExcelImporter) readExcelRows(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no sheets")
	}

	firstSheet := sheets[0]
	i.log.Info("reading excel sheet", slog.String("sheet_name", firstSheet))

	rows, err := f.GetRows(firstSheet)
	if err != nil || len(rows) < 6 {
		return nil, fmt.Errorf("sheet '%s' is empty or missing container headers at row 6", firstSheet)
	}

	return rows, nil
}

// syncProducts cadastra produtos e faz 1 ÚNICO SELECT no banco inteiro no final
func (i *ExcelImporter) syncProducts(rows [][]string) map[string]uint {
	phaseStart := time.Now()
	i.log.Info("Phase 1: Syncing Products to Database...")

	for rowIndex := 6; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		if len(row) < 2 {
			continue
		}

		// CORREÇÃO AQUI: Limpa os espaços e vírgulas logo na Fase 1
		janCode := strings.TrimSpace(row[0])
		productName := strings.TrimSpace(row[1])

		if _, err := strconv.Atoi(janCode); err != nil {
			continue
		}

		product := &models.Product{Name: productName, JanCode: janCode}
		_ = i.productSvc.Create(i.ctx, product)
	}

	i.log.Info("Phase 1: Inserts done. Fetching all products to build cache...")

	var allProducts []models.Product
	i.db.Select("id", "jan_code").Find(&allProducts)

	cache := make(map[string]uint)
	for _, p := range allProducts {
		cache[p.JanCode] = p.ID
	}

	i.log.Info("Phase 1 Complete",
		slog.Int("products_cached", len(cache)),
		slog.String("phase_duration", time.Since(phaseStart).String()),
	)

	return cache
}

// extractContainerData faz o parse de uma única coluna, sanitizando os valores
func (i *ExcelImporter) extractContainerData(rows [][]string, colIndex int, containerName string) ContainerPayload {
	payload := ContainerPayload{
		ContainerName: containerName,
		Items:         make(map[string]int),
	}

	for rowIndex := 6; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		if colIndex >= len(row) {
			continue
		}

		// 1. Pega o valor da célula e limpa
		cellValue := row[colIndex]

		// Remove espaços em branco do começo e do fim (ex: " 5 " vira "5")
		cellValue = strings.TrimSpace(cellValue)
		// Remove vírgulas de formatação de milhar (ex: "1,200" vira "1200")
		cellValue = strings.ReplaceAll(cellValue, ",", "")

		// Se a célula for vazia ou um traço, ignora
		if cellValue == "" || cellValue == "-" {
			continue
		}

		// 2. Converte para número de forma segura
		fQty, err := strconv.ParseFloat(cellValue, 64)
		if err != nil || fQty <= 0 {
			// (Opcional) Se quiser ver o que está falhando, descomente a linha abaixo:
			// i.log.Debug("skipped cell", slog.String("value", cellValue), slog.Int("row", rowIndex))
			continue
		}

		qty := int(fQty) // Converte para inteiro limpo

		janCode := strings.TrimSpace(row[0]) // Limpa o JAN Code também, por segurança
		payload.Items[janCode] += qty
	}

	return payload
}
