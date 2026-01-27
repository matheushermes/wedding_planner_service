package models

import (
	"time"

	"gorm.io/gorm"
)

// Invite representa um convite enviado
type Invite struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	GuestID   uint       `gorm:"not null" json:"guest_id"`
	Guest     Guest      `gorm:"foreignKey:GuestID" json:"guest,omitempty"`
	SentAt    *time.Time `json:"sent_at"`
	SentVia   string     `gorm:"type:varchar(20)" json:"sent_via"` // email, whatsapp
	Template  string     `gorm:"type:text" json:"template"`
	WeddingID uint       `gorm:"not null" json:"wedding_id"`
	Wedding   Wedding    `gorm:"foreignKey:WeddingID" json:"-"`
}
