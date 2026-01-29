package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Wedding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// UserID com índice para queries rápidas de "buscar casamentos do usuário"
	// Performance: Busca O(log n) ao invés de O(n) com full table scan
	UserID uint `gorm:"not null;index:idx_user_weddings" json:"user_id"`

	VenueName         string    `gorm:"size:200" json:"venue_name"`
	VenueAddress      string    `gorm:"type:text" json:"venue_address"`
	EventDate         time.Time `gorm:"index:idx_event_date" json:"event_date"`
	EventTime         string    `gorm:"size:10" json:"event_time"`
	MaxGuests         int       `gorm:"default:0" json:"max_guests"`
	CurrentGuestCount int       `gorm:"default:0" json:"current_guest_count"`
}

// DaysRemaining calcula os dias restantes até o casamento
// Performance: Cálculo em memória, não em query SQL
func (w *Wedding) DaysRemaining() int {
	if w.EventDate.IsZero() {
		return 0
	}
	duration := time.Until(w.EventDate)
	return int(duration.Hours() / 24)
}

// Validate valida todos os campos do wedding
func (w *Wedding) IsValid() error {
	w.normalize()

	if err := w.validateEventDate(); err != nil {
		return err
	}

	if err := w.validateVenueName(); err != nil {
		return err
	}

	if err := w.validateVenueAddress(); err != nil {
		return err
	}

	if err := w.validateEventTime(); err != nil {
		return err
	}

	if err := w.validateMaxGuests(); err != nil {
		return err
	}

	return nil
}

// normalize remove espaços extras dos campos de texto
func (w *Wedding) normalize() {
	w.VenueName = strings.TrimSpace(w.VenueName)
	w.VenueAddress = strings.TrimSpace(w.VenueAddress)
	w.EventTime = strings.TrimSpace(w.EventTime)
}

// validateEventDate valida a data do evento
func (w *Wedding) validateEventDate() error {
	if w.EventDate.IsZero() {
		return errors.New("event date is required")
	}

	// Validação: data não pode ser muito antiga (permite até 1 ano no passado)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	if w.EventDate.Before(oneYearAgo) {
		return errors.New("event date cannot be more than 1 year in the past")
	}

	// Validação: data não pode ser muito futura (máximo 10 anos)
	tenYearsFromNow := time.Now().AddDate(10, 0, 0)
	if w.EventDate.After(tenYearsFromNow) {
		return errors.New("event date cannot be more than 10 years in the future")
	}

	return nil
}

// validateVenueName valida o nome do local
func (w *Wedding) validateVenueName() error {
	if w.VenueName == "" {
		return errors.New("venue name is required")
	}

	if len(w.VenueName) < 3 {
		return errors.New("venue name must be at least 3 characters long")
	}

	if len(w.VenueName) > 200 {
		return errors.New("venue name must not exceed 200 characters")
	}

	return nil
}

// validateVenueAddress valida o endereço do local
func (w *Wedding) validateVenueAddress() error {
	if w.VenueAddress == "" {
		return errors.New("venue address is required")
	}

	if len(w.VenueAddress) < 10 {
		return errors.New("venue address must be at least 10 characters long")
	}

	if len(w.VenueAddress) > 1000 {
		return errors.New("venue address must not exceed 1000 characters")
	}

	return nil
}

// validateEventTime valida o horário do evento
func (w *Wedding) validateEventTime() error {
	if w.EventTime == "" {
		return errors.New("event time is required")
	}

	// Valida formato HH:MM (24h) ou HH:MM AM/PM
	timeRegex24h := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
	timeRegex12h := regexp.MustCompile(`^(0?[1-9]|1[0-2]):[0-5][0-9]\s?(AM|PM|am|pm)$`)

	if !timeRegex24h.MatchString(w.EventTime) && !timeRegex12h.MatchString(w.EventTime) {
		return errors.New("event time must be in format HH:MM or HH:MM AM/PM")
	}

	return nil
}

// validateMaxGuests valida a quantidade máxima de convidados
func (w *Wedding) validateMaxGuests() error {
	if w.MaxGuests < 0 {
		return errors.New("max guests cannot be negative")
	}

	if w.MaxGuests > 10000 {
		return errors.New("max guests cannot exceed 10,000")
	}

	// Validação de consistência: CurrentGuestCount não pode exceder MaxGuests
	if w.CurrentGuestCount > w.MaxGuests {
		return errors.New("current guest count cannot exceed max guests")
	}

	return nil
}
