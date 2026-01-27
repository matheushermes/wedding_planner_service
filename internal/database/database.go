package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/matheushermes/wedding_planner_service/configs"
	"github.com/matheushermes/wedding_planner_service/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB conecta ao banco com retry e configurações otimizadas
func ConnectDB() error {
	dsn := configs.DATABASE_URL
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL não configurada")
	}

	// Logger customizado para não expor credenciais
	var logLevel logger.LogLevel
	if configs.ENV == "production" {
		logLevel = logger.Error // Apenas erros em produção
	} else {
		logLevel = logger.Info
	}

	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Log queries lentas
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  configs.ENV != "production",
		},
	)

	// Retry com backoff exponencial
	maxRetries := 5
	var err error

	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:                                   customLogger,
			DisableForeignKeyConstraintWhenMigrating: false,
			PrepareStmt:                              true, // Prepared statements para performance
		})

		if err == nil {
			break
		}

		waitTime := time.Duration(i+1) * 2 * time.Second
		log.Printf("⚠️  Tentativa %d/%d falhou. Aguardando %v... (DSN: %s)", i+1, maxRetries, waitTime, configs.MaskDSN(dsn))
		time.Sleep(waitTime)
	}

	if err != nil {
		return fmt.Errorf("falha ao conectar após %d tentativas: %w", maxRetries, err)
	}

	// Configurações de pool otimizadas para produção
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("erro ao obter *sql.DB: %w", err)
	}

	// Configurações ajustadas para carga
	sqlDB.SetMaxIdleConns(10)                   // Conexões idle
	sqlDB.SetMaxOpenConns(configs.MAX_DB_CONNS) // Máximo de conexões abertas
	sqlDB.SetConnMaxLifetime(time.Hour)         // Tempo máximo de vida de conexão
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)  // Tempo máximo idle

	// Testa a conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("falha no ping do banco: %w", err)
	}

	log.Println("✅ Conexão com banco de dados estabelecida com sucesso!")
	return nil
}

// InitializeDatabase inicializa o banco e executa migrações
func InitializeDatabase() {
	if err := ConnectDB(); err != nil {
		log.Fatalf("❌ Erro fatal ao conectar ao banco: %v", err)
	}

	// Executa migrações em desenvolvimento e staging
	if configs.ENV != "production" {
		log.Println("🔄 Executando migrações automáticas...")
		if err := MigrateDB(
			&models.User{},
			&models.Wedding{},
			&models.Fundraising{},
			&models.Guest{},
			&models.Invite{},
		); err != nil {
			log.Fatalf("❌ Erro ao executar migrações: %v", err)
		}
		log.Println("✅ Migrações concluídas!")
	} else {
		log.Println("ℹ️  Modo produção: migrações automáticas desabilitadas")
	}
}

// CloseDatabase fecha a conexão com o banco gracefully
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		log.Println("🔒 Fechando conexões com o banco...")
		return sqlDB.Close()
	}
	return nil
}