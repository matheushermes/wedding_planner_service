package controllers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matheushermes/wedding_planner_service/internal/auth"
	"github.com/matheushermes/wedding_planner_service/internal/database"
	"github.com/matheushermes/wedding_planner_service/internal/models"
	"github.com/matheushermes/wedding_planner_service/internal/repository"
	"github.com/matheushermes/wedding_planner_service/internal/security"
	"gorm.io/gorm"
)

// Constantes de configuração
const (
	maxRequestBodySize = 1 << 20 // 1MB
	timingAttackDelay  = 100 * time.Millisecond
)

// Response structs padronizadas para consistência da API
type userResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PartnerName string    `json:"partner_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int64        `json:"expires_in"` // em segundos
	User      userResponse `json:"user"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// RegisterUser registra um novo usuário no sistema
func RegisterUser(c *gin.Context) {
	var user models.User

	// Proteção contra DoS (limita tamanho do body)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponse{
			Error: "invalid request data",
		})
		return
	}

	// Validação das regras de negócio
	if err := user.IsValid("register"); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewUserRepository(database.DB)
	if err := repo.Create(&user); err != nil {
		log.Printf("[ERROR] Failed to create user: %v", err)

		// Tratamento de erro de duplicação (race condition entre check e insert)
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, errorResponse{
				Error: "unable to register user, please check your data",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to register user at this time",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user": userResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PartnerName: user.PartnerName,
			CreatedAt:   user.CreatedAt,
		},
	})
}

// Login autentica um usuário e retorna um token JWT
func Login(c *gin.Context) {
	var loginReq models.LoginRequest

	// Proteção contra DoS
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponse{
			Error: "invalid request data",
		})
		return
	}

	// Busca usuário no banco
	repo := repository.NewUserRepository(database.DB)
	user, err := repo.FindByEmail(loginReq.Email)
	if err != nil {
		// Delay constante para prevenir timing attacks (impede enumeração de usuários)
		time.Sleep(timingAttackDelay)

		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "invalid email or password",
		})
		return
	}

	// Verifica a senha
	if err := security.CheckPassword(user.PasswordHash, loginReq.Password); err != nil {
		// Delay constante para prevenir timing attacks
		time.Sleep(timingAttackDelay)

		// Log de tentativa falha (detecção de brute force)
		log.Printf("[SECURITY] Failed login attempt for email: %s from IP: %s", loginReq.Email, c.ClientIP())

		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "invalid email or password",
		})
		return
	}

	// Gera token JWT
	token, err := auth.CreateToken(user.ID, user.Email)
	if err != nil {
		log.Printf("[ERROR] Failed to create token for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to complete authentication",
		})
		return
	}

	// Log de login bem-sucedido (auditoria)
	log.Printf("[INFO] Successful login for user %d (%s) from IP: %s", user.ID, user.Email, c.ClientIP())

	// Resposta estruturada
	c.JSON(http.StatusOK, loginResponse{
		Token:     token,
		ExpiresIn: int64(auth.TokenExpirationTime.Seconds()),
		User: userResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PartnerName: user.PartnerName,
			CreatedAt:   user.CreatedAt,
		},
	})
}

// GetProfile retorna o perfil do usuário autenticado
func GetProfile(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	repo := repository.NewUserRepository(database.DB)
	user, err := repo.FindByID(userID.(uint))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{
				Error: "user not found",
			})
			return
		}

		log.Printf("[ERROR] Failed to fetch user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to fetch user profile",
		})
		return
	}

	c.JSON(http.StatusOK, userResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PartnerName: user.PartnerName,
		CreatedAt:   user.CreatedAt,
	})
}

// UpdateProfile atualiza o perfil do usuário autenticado
func UpdateProfile(c *gin.Context) {
	userID, err := auth.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: err.Error(),
		})
		return
	}

	var updateData struct {
		Name        string `json:"name" binding:"omitempty,min=2,max=100"`
		PartnerName string `json:"partner_name" binding:"omitempty,max=100"`
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request data",
		})
		return
	}

	repo := repository.NewUserRepository(database.DB)
	user, err := repo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "user not found",
		})
		return
	}

	// Atualiza apenas campos fornecidos (PATCH behavior)
	if updateData.Name != "" {
		user.Name = strings.TrimSpace(updateData.Name)
	}
	if updateData.PartnerName != "" {
		user.PartnerName = strings.TrimSpace(updateData.PartnerName)
	}

	if err := repo.Update(user); err != nil {
		log.Printf("[ERROR] Failed to update user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
		"user": userResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PartnerName: user.PartnerName,
			CreatedAt:   user.CreatedAt,
		},
	})
}

// DeleteUser deleta a conta do usuário autenticado (soft delete)
func DeleteUser(c *gin.Context) {
	userID, err := auth.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewUserRepository(database.DB)
	if err := repo.Delete(userID); err != nil {
		log.Printf("[ERROR] Failed to delete user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to delete user account",
		})
		return
	}

	// Log de auditoria
	log.Printf("[INFO] User %d deleted account from IP: %s", userID, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "user account deleted successfully",
	})
}
