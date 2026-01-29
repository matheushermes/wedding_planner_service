package repository

import (
	"errors"

	"github.com/matheushermes/wedding_planner_service/internal/models"
	"gorm.io/gorm"
)

// UserRepository encapsula as operações de banco de dados para usuários
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository cria uma nova instância do UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create cria um novo usuário no banco de dados
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail busca um usuário pelo email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByID busca um usuário pelo ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update atualiza os dados de um usuário
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete deleta um usuário (soft delete)
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// HardDelete deleta um usuário permanentemente do banco de dados
func (r *UserRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.User{}, id).Error
}
