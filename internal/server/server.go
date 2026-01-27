package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matheushermes/wedding_planner_service/configs"
	"github.com/matheushermes/wedding_planner_service/internal/database"
	"github.com/matheushermes/wedding_planner_service/internal/server/routes"
)

type Server struct {
	port   string
	server *gin.Engine
}

// NewServer cria nova instância do servidor
func NewServer() Server {
	// Define modo do Gin baseado no ambiente
	gin.SetMode(configs.GIN_MODE)

	return Server{
		port:   configs.PORT,
		server: gin.Default(),
	}
}

// RunServer inicia o servidor com graceful shutdown
func (s *Server) RunServer() error {
	router := routes.ConfigRoutes(s.server)

	// Configuração do servidor HTTP
	srv := &http.Server{
		Addr:           ":" + s.port,
		Handler:        router,
		ReadTimeout:    time.Duration(configs.READ_TIMEOUT_SECS) * time.Second,
		WriteTimeout:   time.Duration(configs.WRITE_TIMEOUT_SECS) * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Canal para erros do servidor
	serverErrors := make(chan error, 1)

	// Inicia servidor em goroutine
	go func() {
		log.Printf("🚀 Servidor iniciado em http://localhost:%s (ENV: %s)", s.port, configs.ENV)
		serverErrors <- srv.ListenAndServe()
	}()

	// Canal para sinais de shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Aguarda erro ou sinal de shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("erro ao iniciar servidor: %w", err)

	case sig := <-shutdown:
		log.Printf("🛑 Sinal de shutdown recebido: %v", sig)

		// Contexto com timeout para graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fecha servidor HTTP
		log.Println("🔄 Encerrando servidor HTTP...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("⚠️  Erro no shutdown do servidor: %v", err)
			if err := srv.Close(); err != nil {
				return fmt.Errorf("erro ao forçar fechamento: %w", err)
			}
		}

		// Fecha conexões do banco
		log.Println("🔄 Fechando conexões com o banco...")
		if err := database.CloseDatabase(); err != nil {
			log.Printf("⚠️  Erro ao fechar banco: %v", err)
		}

		log.Println("✅ Shutdown concluído com sucesso!")
	}

	return nil
}