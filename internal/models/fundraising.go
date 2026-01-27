package models

import (
	"time"

	"gorm.io/gorm"
)

// Fundraising representa uma arrecadação
type Fundraising struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	WeddingID       uint              `gorm:"not null" json:"wedding_id"`
	Wedding         Wedding           `gorm:"foreignKey:WeddingID" json:"-"`
	Type            FundraisingType   `gorm:"type:varchar(20);not null" json:"type"`
	Amount          float64           `gorm:"not null" json:"amount"`
	Date            time.Time         `json:"date"`
	Observation     string            `gorm:"type:text" json:"observation"`
	DonorName       string            `json:"donor_name"` // nome de quem doou
}

// FundraisingType representa os tipos de arrecadação
type FundraisingType string

const (
	FundraisingTypeGift   FundraisingType = "gift"
	FundraisingTypeTie    FundraisingType = "tie"       // Gravata
	FundraisingTypeShoe   FundraisingType = "shoe"      // Sapatinho
)
