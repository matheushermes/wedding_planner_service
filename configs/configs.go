package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	PORT               string
	DATABASE_URL       string
	ENV                string
	GIN_MODE           string
	MAX_DB_CONNS       int
	READ_TIMEOUT_SECS  int
	WRITE_TIMEOUT_SECS int
	JWT_SECRET         []byte
)

// LoadEnv carrega e valida variáveis de ambiente
func LoadEnv() {
	// Carrega .env apenas em desenvolvimento
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("⚠️  Arquivo .env não encontrado, usando variáveis do sistema")
		}
	}

	// Ambiente
	ENV = getEnv("ENV", "development")

	// Porta
	PORT = getEnv("PORT", "8080")

	// Gin Mode
	GIN_MODE = getEnv("GIN_MODE", "debug")
	if ENV == "production" && GIN_MODE == "debug" {
		GIN_MODE = "release"
		log.Println("⚠️  GIN_MODE alterado para 'release' em ambiente de produção")
	}

	// Database URL - CRÍTICO
	DATABASE_URL = os.Getenv("DATABASE_URL")
	if DATABASE_URL == "" {
		log.Fatal("❌ DATABASE_URL não definida")
	}

	// Valida formato básico da DSN
	if !strings.Contains(DATABASE_URL, "@") || !strings.Contains(DATABASE_URL, "tcp(") {
		log.Fatal("❌ DATABASE_URL inválida. Formato esperado: user:pass@tcp(host:port)/dbname?params")
	}

	// JWT Secret - CRÍTICO
	JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
	if len(JWT_SECRET) == 0 {
		log.Fatal("❌ JWT_SECRET não definida")
	}

	// Configurações de performance
	MAX_DB_CONNS = getEnvInt("MAX_DB_CONNS", 100)
	READ_TIMEOUT_SECS = getEnvInt("READ_TIMEOUT_SECS", 30)
	WRITE_TIMEOUT_SECS = getEnvInt("WRITE_TIMEOUT_SECS", 30)

	log.Printf("✅ Configurações carregadas: ENV=%s, PORT=%s, GIN_MODE=%s", ENV, PORT, GIN_MODE)
}

// getEnv retorna variável de ambiente ou valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retorna variável de ambiente int ou valor padrão
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		log.Printf("⚠️  %s inválido, usando padrão: %d", key, defaultValue)
	}
	return defaultValue
}

// MaskDSN mascara credenciais da DSN para logs seguros
func MaskDSN(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx > 0 {
		return fmt.Sprintf("***%s", dsn[idx:])
	}
	return "***"
}
