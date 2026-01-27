package models

import (
	"time"

	"gorm.io/gorm"
)

// Guest representa um convidado
type Guest struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	FullName     string       `gorm:"not null" json:"full_name"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email"`
	InviteStatus InviteStatus `gorm:"type:varchar(20);default:'pending'" json:"invite_status"`
	MaxGuests    int          `gorm:"default:1" json:"max_guests"` // número máximo de convidados que essa pessoa pode trazer
	WeddingID    uint         `gorm:"not null" json:"wedding_id"`
	Wedding      Wedding      `gorm:"foreignKey:WeddingID" json:"-"`
}

// InviteStatus representa os possíveis status de convite
type InviteStatus string

const (
	InviteStatusPending   InviteStatus = "pending"
	InviteStatusSent      InviteStatus = "sent"
	InviteStatusConfirmed InviteStatus = "confirmed"
	InviteStatusDeclined  InviteStatus = "declined"
)
