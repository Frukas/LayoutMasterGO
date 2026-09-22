package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// NewConnection estabelece a conexão com o banco sem depender de nenhuma implementação específica.
func NewConnection(dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco de dados: %w", err)
	}

	return db, nil
}

// NewSQLiteDialector cria o dialector específico para o SQLite com WAL mode ativado.
func NewSQLiteDialector(dbPath string) gorm.Dialector {
	return sqlite.Open(dbPath + "?_journal_mode=WAL")
}
