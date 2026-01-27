package database

import (
	"fmt"
	"log"
)

// MigrateDB executa migrações com tratamento de erro
func MigrateDB(models ...interface{}) error {
	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("erro ao migrar %T: %w", model, err)
		}
		log.Printf("  ✅ Migração completa: %T", model)
	}
	return nil
}