package main

import (
	"log"

	_ "github.com/matheushermes/wedding_planner_service/init"
	"github.com/matheushermes/wedding_planner_service/internal/database"
	"github.com/matheushermes/wedding_planner_service/internal/server"
)

func main() {
	log.Println("💒 Iniciando Wedding Planner Service...")

	// Inicializa banco de dados
	database.InitializeDatabase()

	// Cria servidor
	appServer := server.NewServer()

	// Inicia servidor (com graceful shutdown)
	if err := appServer.RunServer(); err != nil {
		log.Fatalf("❌ Erro fatal: %v", err)
	}
}