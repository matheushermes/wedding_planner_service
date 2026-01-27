package models

import (
	"time"

	"gorm.io/gorm"
)

// Wedding representa os dados principais do casamento
type Wedding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID            uint      `gorm:"not null" json:"user_id"`
	VenueName         string    `json:"venue_name"`
	VenueAddress      string    `gorm:"type:text" json:"venue_address"`
	EventDate         time.Time `json:"event_date"`
	EventTime         string    `json:"event_time"`
	MaxGuests         int       `gorm:"default:0" json:"max_guests"`
	CurrentGuestCount int       `gorm:"default:0" json:"current_guest_count"`
}

// DaysRemaining calcula os dias restantes até o casamento
func (w *Wedding) DaysRemaining() int {
	if w.EventDate.IsZero() {
		return 0
	}
	duration := time.Until(w.EventDate)
	return int(duration.Hours() / 24)
}
