package importer

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"layoutmastergo/internal/models"
)

// processContainers atua como o Produtor, lendo as colunas e alimentando o Canal
// processContainers atua como o Produtor, lendo as colunas e alimentando o Canal
func (i *ExcelImporter) processContainers(rows [][]string, cache map[string]uint) {
	phaseStart := time.Now()
	i.log.Info("Phase 2: Starting Worker Pool for Containers...")

	headerRow := rows[5] // Linha 6 (Índice 5)

	// Descobre a largura máxima real da planilha inteira para evitar "index out of range"
	numColumns := 0
	for _, r := range rows {
		if len(r) > numColumns {
			numColumns = len(r)
		}
	}

	jobs := make(chan ContainerPayload, numColumns)
	var wg sync.WaitGroup

	const numWorkers = 3
	i.startWorkerPool(numWorkers, jobs, cache, &wg)

	totalDispatched := 0
	const startColumnIndex = 13 // 13 = Coluna N (Início real dos contêineres)

	// NOVO: Contador de colunas vazias seguidas
	consecutiveEmpty := 0

	for colIndex := startColumnIndex; colIndex < numColumns; colIndex++ {
		containerName := ""
		if colIndex < len(headerRow) {
			// Limpa espaços normais e também espaços "inquebráveis" (NBSP) nativos do Excel
			rawName := strings.ReplaceAll(headerRow[colIndex], "\u00A0", " ")
			containerName = strings.TrimSpace(rawName)
		}

		// Se a coluna estiver vazia, incrementa o contador
		if containerName == "" {
			consecutiveEmpty++

			// Se achar 3 colunas vazias SEGUIDAS, significa que a tabela acabou!
			if consecutiveEmpty >= 2 {
				i.log.Info("End of matrix detected. Stopping scan.", slog.Int("stopped_at_col", colIndex))
				break // Aborta o loop e ignora o resto da planilha fantasma
			}
			continue
		}

		// Se achou um nome válido, zera o contador de vazios!
		consecutiveEmpty = 0

		// Ignora o Totalizador ou se for apenas um número avulso
		if strings.ToUpper(containerName) == "TOTAL" {
			continue
		}
		if _, err := strconv.Atoi(containerName); err == nil {
			continue
		}

		i.log.Info("container parsed", slog.Int("col", colIndex), slog.String("name", containerName))

		payload := i.extractContainerData(rows, colIndex, containerName)
		if len(payload.Items) > 0 {
			jobs <- payload
			totalDispatched++
		}
	}

	i.log.Info("Producer dispatched all container jobs", slog.Int("total_containers_dispatched", totalDispatched))

	close(jobs) // Finaliza o canal, avisando os workers que não há mais pacotes
	wg.Wait()   // Aguarda o término de todas as goroutines

	i.log.Info("Phase 2 Complete", slog.String("phase_duration", time.Since(phaseStart).String()))
}

func (i *ExcelImporter) startWorkerPool(workers int, jobs <-chan ContainerPayload, cache map[string]uint, wg *sync.WaitGroup) {
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go i.worker(w, jobs, cache, wg)
	}
}

func (i *ExcelImporter) worker(workerID int, jobs <-chan ContainerPayload, cache map[string]uint, wg *sync.WaitGroup) {
	defer wg.Done()

	for payload := range jobs {
		i.log.Info("worker started job", slog.Int("worker_id", workerID), slog.String("container", payload.ContainerName))

		// 1. Cria ou Busca o ID do Contêiner
		container := &models.Container{Name: payload.ContainerName, Status: "Active"}
		err := i.containerSvc.Create(i.ctx, container)

		if err != nil || container.ID == 0 {
			var existing models.Container
			if errFetch := i.db.Where("name = ?", payload.ContainerName).First(&existing).Error; errFetch == nil {
				container.ID = existing.ID
			} else {
				i.log.Error("failed to resolve container ID", slog.String("container", payload.ContainerName))
				continue
			}
		}

		// 2. Vincula os itens do contêiner
		for janCode, qty := range payload.Items {
			productID, exists := cache[janCode]
			if !exists {
				continue
			}

			if err := i.containerSvc.AddProduct(i.ctx, container.ID, productID, qty); err != nil {
				i.log.Error("failed to add product to container",
					slog.String("container", payload.ContainerName),
					slog.String("jan_code", janCode),
					slog.String("error", err.Error()),
				)
			}
		}

		i.log.Info("worker finished job", slog.Int("worker_id", workerID), slog.String("container", payload.ContainerName))
	}
}
