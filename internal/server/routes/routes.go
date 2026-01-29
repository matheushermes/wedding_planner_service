package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matheushermes/wedding_planner_service/configs"
	"github.com/matheushermes/wedding_planner_service/internal/controllers"
	"github.com/matheushermes/wedding_planner_service/internal/server/middlewares"
)

// ConfigRoutes configura todas as rotas da API
func ConfigRoutes(router *gin.Engine) *gin.Engine {
	// Middleware de recovery para evitar crash
	router.Use(gin.Recovery())

	// Middleware de CORS para produção
	if configs.ENV == "production" {
		router.Use(corsMiddleware())
	}

	// Grupo principal da API
	api := router.Group("/api/v1")
	{
		// Health detalhado
		health := api.Group("/health")
		{
			health.GET("/status", healthCheck)
		}

		// User - Autenticação
		user := api.Group("/user")
		{
			// 🌐 públicas
			user.POST("/register", controllers.RegisterUser)
			user.POST("/login", controllers.Login)

			// 🔐 privadas
			user.Use(middlewares.AuthMiddleware())
			{
				user.GET("/profile", controllers.GetProfile)
				user.PATCH("/update", controllers.UpdateProfile)
				user.DELETE("/delete", controllers.DeleteUser)
				user.POST("/logout", nil)
			}
		}

		// Wedding - Dados do Casamento
		weddings := api.Group("/weddings", middlewares.AuthMiddleware())
		{
			weddings.POST("/", controllers.CreateWedding)
			weddings.GET("/", controllers.GetWeddings)
			weddings.GET("/:id", controllers.GetWedding)
			weddings.PUT("/:id", controllers.UpdateWedding)
			weddings.DELETE("/:id", controllers.DeleteWedding)

			// Recursos aninhados dentro do wedding
			wedding := weddings.Group("/:id")
			{
				// Contagem regressiva
				wedding.GET("/countdown", controllers.GetCountdown)

				// Guests - Módulo de Convidados
				guests := wedding.Group("/guests")
				{
					guests.POST("", nil)            // TODO: Implementar controller - Cadastrar convidado
					guests.POST("/batch", nil)      // TODO: Implementar controller - Cadastrar convidados em lote
					guests.GET("", nil)             // TODO: Implementar controller - Listar todos os convidados
					guests.GET("/stats", nil)       // TODO: Implementar controller - Estatísticas de convidados
					guests.GET("/:guestId", nil)    // TODO: Implementar controller - Obter convidado específico
					guests.PUT("/:guestId", nil)    // TODO: Implementar controller - Editar convidado
					guests.DELETE("/:guestId", nil) // TODO: Implementar controller - Remover convidado
				}

				// Invites - Módulo de Convites Automáticos
				invites := wedding.Group("/invites")
				{
					invites.POST("", nil)                  // TODO: Implementar controller - Criar convite
					invites.GET("", nil)                   // TODO: Implementar controller - Listar convites
					invites.GET("/:inviteId", nil)         // TODO: Implementar controller - Obter convite específico
					invites.PUT("/:inviteId", nil)         // TODO: Implementar controller - Atualizar convite
					invites.POST("/:inviteId/send", nil)   // TODO: Implementar controller - Enviar convite
					invites.POST("/:inviteId/resend", nil) // TODO: Implementar controller - Reenviar convite
				}

				// Budget - Módulo de Orçamento
				budget := wedding.Group("/budget")
				{
					budget.POST("", nil)        // TODO: Implementar controller - Definir orçamento
					budget.GET("", nil)         // TODO: Implementar controller - Obter orçamento
					budget.PUT("", nil)         // TODO: Implementar controller - Atualizar orçamento
					budget.GET("/summary", nil) // TODO: Implementar controller - Resumo do orçamento
				}

				// Expenses - Gastos
				expenses := wedding.Group("/expenses")
				{
					expenses.POST("", nil)                    // TODO: Implementar controller - Cadastrar gasto
					expenses.GET("", nil)                     // TODO: Implementar controller - Listar gastos
					expenses.GET("/by-category", nil)         // TODO: Implementar controller - Listar gastos por categoria
					expenses.GET("/:expenseId", nil)          // TODO: Implementar controller - Obter gasto específico
					expenses.PUT("/:expenseId", nil)          // TODO: Implementar controller - Atualizar gasto
					expenses.DELETE("/:expenseId", nil)       // TODO: Implementar controller - Deletar gasto
					expenses.PATCH("/:expenseId/status", nil) // TODO: Implementar controller - Marcar como pago/previsto
				}

				// Fundraising - Módulo de Arrecadações
				fundraising := wedding.Group("/fundraising")
				{
					fundraising.POST("", nil)                  // TODO: Implementar controller - Registrar arrecadação
					fundraising.GET("", nil)                   // TODO: Implementar controller - Listar arrecadações
					fundraising.GET("/summary", nil)           // TODO: Implementar controller - Resumo de arrecadações
					fundraising.GET("/by-type", nil)           // TODO: Implementar controller - Arrecadações por tipo
					fundraising.GET("/:fundraisingId", nil)    // TODO: Implementar controller - Obter arrecadação específica
					fundraising.PUT("/:fundraisingId", nil)    // TODO: Implementar controller - Atualizar arrecadação
					fundraising.DELETE("/:fundraisingId", nil) // TODO: Implementar controller - Deletar arrecadação
				}
			}
		}
	}

	return router
}

// healthCheck handler de health check
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"env":    configs.ENV,
	})
}

// corsMiddleware middleware de CORS para produção
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
