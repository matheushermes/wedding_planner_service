package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/matheushermes/wedding_planner_service/internal/auth"
)

// AuthMiddleware é um middleware para proteger rotas que requerem autenticação
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verifica o token
		if err := auth.VerifyToken(c); err != nil {
			c.JSON(401, gin.H{
				"error": err.Error(),
			})
			c.Abort() // Impede que handlers subsequentes sejam executados
			return
		}
		
		userID, err := auth.ExtractUserID(c)
		if err != nil {
			c.JSON(401, gin.H{
				"error": "failed to extract user information",
			})
			c.Abort()
			return
		}

		// Armazena o user_id no contexto para uso nos handlers
		c.Set("user_id", userID)

		// Continua para o próximo handler
		c.Next()
	}
}