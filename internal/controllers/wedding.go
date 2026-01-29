package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matheushermes/wedding_planner_service/internal/database"
	"github.com/matheushermes/wedding_planner_service/internal/models"
	"github.com/matheushermes/wedding_planner_service/internal/repository"
)

// weddingResponse representa a resposta padronizada de wedding
type weddingResponse struct {
	ID                uint      `json:"id"`
	UserID            uint      `json:"user_id"`
	VenueName         string    `json:"venue_name"`
	VenueAddress      string    `json:"venue_address"`
	EventDate         time.Time `json:"event_date"`
	EventTime         string    `json:"event_time"`
	MaxGuests         int       `json:"max_guests"`
	CurrentGuestCount int       `json:"current_guest_count"`
	DaysRemaining     int       `json:"days_remaining"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// weddingListResponse retorna dados resumidos para listagem
// Performance: Reduz payload de resposta excluindo campos desnecessários
type weddingListResponse struct {
	ID            uint      `json:"id"`
	VenueName     string    `json:"venue_name"`
	EventDate     time.Time `json:"event_date"`
	EventTime     string    `json:"event_time"`
	MaxGuests     int       `json:"max_guests"`
	GuestCount    int       `json:"guest_count"`
	DaysRemaining int       `json:"days_remaining"`
}

// countdownResponse retorna apenas contagem regressiva
type countdownResponse struct {
	EventDate     time.Time `json:"event_date"`
	DaysRemaining int       `json:"days_remaining"`
	Status        string    `json:"status"` // upcoming, today, past
}

// CreateWedding cria um novo casamento para o usuário autenticado
func CreateWedding(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	var wedding models.Wedding
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)

	if err := c.ShouldBindJSON(&wedding); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request data",
		})
		return
	}

	// Associa o casamento ao usuário autenticado
	// Segurança: Impede que usuário crie casamento para outro user_id
	wedding.UserID = userID.(uint)

	// Validações de negócio no model
	if err := wedding.IsValid(); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)

	// Performance: Uma única operação de INSERT no banco
	if err := repo.Create(&wedding); err != nil {
		log.Printf("[ERROR] Failed to create wedding for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to create wedding",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "wedding created successfully",
		"wedding": toWeddingResponse(&wedding),
	})
}

// GetWeddings lista todos os casamentos do usuário autenticado
func GetWeddings(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)

	// Performance: Query otimizada com índice em user_id + ordenação
	weddings, err := repo.FindByUserID(userID.(uint))
	if err != nil {
		log.Printf("[ERROR] Failed to fetch weddings for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to fetch weddings",
		})
		return
	}

	// Performance: Retorna array vazio ao invés de null se não houver dados
	// Facilita parsing no frontend e reduz bugs
	if len(weddings) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"weddings": []weddingListResponse{},
			"count":    0,
		})
		return
	}

	// Performance: Mapeia para response reduzido (menos dados na rede)
	response := make([]weddingListResponse, len(weddings))
	for i, w := range weddings {
		response[i] = weddingListResponse{
			ID:            w.ID,
			VenueName:     w.VenueName,
			EventDate:     w.EventDate,
			EventTime:     w.EventTime,
			MaxGuests:     w.MaxGuests,
			GuestCount:    w.CurrentGuestCount,
			DaysRemaining: w.DaysRemaining(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"weddings": response,
		"count":    len(response),
	})
}

// GetWedding retorna detalhes de um casamento específico
func GetWedding(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	// Extrai e valida ID do casamento da URL
	weddingID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)

	// Performance: Query com índice composto (id + user_id)
	// Segurança: Verifica ownership em uma única query
	wedding, err := repo.FindByIDAndUserID(weddingID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wedding": toWeddingResponse(wedding),
	})
}

// UpdateWedding atualiza os dados de um casamento
func UpdateWedding(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	weddingID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)

	// Busca e valida ownership
	wedding, err := repo.FindByIDAndUserID(weddingID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: err.Error(),
		})
		return
	}

	// Estrutura para atualização parcial
	var updateData struct {
		VenueName    *string    `json:"venue_name"`
		VenueAddress *string    `json:"venue_address"`
		EventDate    *time.Time `json:"event_date"`
		EventTime    *string    `json:"event_time"`
		MaxGuests    *int       `json:"max_guests"`
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request data",
		})
		return
	}

	// Atualiza apenas campos fornecidos (PATCH behavior)
	// Performance: Evita sobrescrever dados desnecessariamente
	if updateData.VenueName != nil {
		wedding.VenueName = *updateData.VenueName
	}
	if updateData.VenueAddress != nil {
		wedding.VenueAddress = *updateData.VenueAddress
	}
	if updateData.EventDate != nil {
		wedding.EventDate = *updateData.EventDate
	}
	if updateData.EventTime != nil {
		wedding.EventTime = *updateData.EventTime
	}
	if updateData.MaxGuests != nil {
		wedding.MaxGuests = *updateData.MaxGuests
	}

	// Validações após atualização (normalize é chamado dentro do IsValid)
	if err := wedding.IsValid(); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	// Performance: GORM otimiza UPDATE apenas dos campos alterados
	if err := repo.Update(wedding); err != nil {
		log.Printf("[ERROR] Failed to update wedding %d: %v", weddingID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to update wedding",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "wedding updated successfully",
		"wedding": toWeddingResponse(wedding),
	})
}

// DeleteWedding remove um casamento (soft delete)
func DeleteWedding(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	weddingID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)

	// Verifica ownership antes de deletar
	wedding, err := repo.FindByIDAndUserID(weddingID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: err.Error(),
		})
		return
	}

	// Performance: Soft delete é mais rápido que DELETE físico
	// Mantém integridade referencial com guests, budget, etc
	if err := repo.Delete(weddingID); err != nil {
		log.Printf("[ERROR] Failed to delete wedding %d: %v", weddingID, err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "unable to delete wedding",
		})
		return
	}

	log.Printf("[INFO] User %d deleted wedding %d (%s)", userID, weddingID, wedding.VenueName)

	c.JSON(http.StatusOK, gin.H{
		"message": "wedding deleted successfully",
	})
}

// GetCountdown retorna contagem regressiva até o casamento
func GetCountdown(c *gin.Context) {
	// Pega userID do contexto (colocado pelo AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errorResponse{
			Error: "authentication required",
		})
		return
	}

	weddingID, err := parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
		return
	}

	repo := repository.NewWeddingRepository(database.DB)
	wedding, err := repo.FindByIDAndUserID(weddingID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: err.Error(),
		})
		return
	}

	// Calcula status baseado na data
	daysRemaining := wedding.DaysRemaining()
	status := "upcoming"
	if daysRemaining < 0 {
		status = "past"
	} else if daysRemaining == 0 {
		status = "today"
	}

	c.JSON(http.StatusOK, countdownResponse{
		EventDate:     wedding.EventDate,
		DaysRemaining: daysRemaining,
		Status:        status,
	})
}

// toWeddingResponse converte model para response
// Performance: Centraliza lógica de conversão evitando duplicação
func toWeddingResponse(w *models.Wedding) weddingResponse {
	return weddingResponse{
		ID:                w.ID,
		UserID:            w.UserID,
		VenueName:         w.VenueName,
		VenueAddress:      w.VenueAddress,
		EventDate:         w.EventDate,
		EventTime:         w.EventTime,
		MaxGuests:         w.MaxGuests,
		CurrentGuestCount: w.CurrentGuestCount,
		DaysRemaining:     w.DaysRemaining(),
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.UpdatedAt,
	}
}

// parseIDParam extrai e valida ID da URL
// Performance: Função reutilizável evita código duplicado
func parseIDParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("invalid ID parameter")
	}
	return uint(id), nil
}
