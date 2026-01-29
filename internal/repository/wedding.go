package repository

import (
	"errors"

	"github.com/matheushermes/wedding_planner_service/internal/models"
	"gorm.io/gorm"
)

type WeddingRepository struct {
	db *gorm.DB
}

func NewWeddingRepository(db *gorm.DB) *WeddingRepository {
	return &WeddingRepository{db: db}
}

// Create cria um novo casamento
// Performance: Usa apenas uma operação de INSERT no banco
func (r *WeddingRepository) Create(wedding *models.Wedding) error {
	return r.db.Create(wedding).Error
}

// FindByID busca um casamento pelo ID
// Performance: Usa índice de primary key para busca O(log n)
func (r *WeddingRepository) FindByID(id uint) (*models.Wedding, error) {
	var wedding models.Wedding
	err := r.db.First(&wedding, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wedding not found")
		}
		return nil, err
	}
	return &wedding, nil
}

// FindByUserID lista todos os casamentos de um usuário
// Performance: Usa índice em user_id para busca eficiente
// Ordenação por event_date para mostrar próximos eventos primeiro
func (r *WeddingRepository) FindByUserID(userID uint) ([]models.Wedding, error) {
	var weddings []models.Wedding
	err := r.db.Where("user_id = ?", userID).
		Order("event_date ASC").
		Find(&weddings).Error
	if err != nil {
		return nil, err
	}
	return weddings, nil
}

// FindByIDAndUserID busca um casamento específico de um usuário
// Performance: Usa índices compostos para verificação de ownership em O(log n)
// Segurança: Garante que usuário só acesse seus próprios dados
func (r *WeddingRepository) FindByIDAndUserID(weddingID, userID uint) (*models.Wedding, error) {
	var wedding models.Wedding
	err := r.db.Where("id = ? AND user_id = ?", weddingID, userID).First(&wedding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wedding not found or access denied")
		}
		return nil, err
	}
	return &wedding, nil
}

// Update atualiza os dados de um casamento
// Performance: Usa Save() que otimiza apenas campos alterados
func (r *WeddingRepository) Update(wedding *models.Wedding) error {
	return r.db.Save(wedding).Error
}

// Delete remove um casamento (soft delete)
// Performance: Soft delete é mais rápido que DELETE físico e mantém integridade referencial
func (r *WeddingRepository) Delete(id uint) error {
	return r.db.Delete(&models.Wedding{}, id).Error
}

// HardDelete remove permanentemente um casamento
// Use apenas quando necessário limpar completamente os dados
func (r *WeddingRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.Wedding{}, id).Error
}

// CountByUserID retorna a quantidade de casamentos de um usuário
// Performance: COUNT é otimizado pelo GORM, mais rápido que buscar todos registros
func (r *WeddingRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Wedding{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// UpdateGuestCount atualiza apenas o contador de convidados
// Performance: UPDATE de um único campo é mais eficiente que Save() completo
func (r *WeddingRepository) UpdateGuestCount(weddingID uint, count int) error {
	return r.db.Model(&models.Wedding{}).
		Where("id = ?", weddingID).
		Update("current_guest_count", count).Error
}
