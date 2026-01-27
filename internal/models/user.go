package models

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/matheushermes/wedding_planner_service/internal/security"
	"gorm.io/gorm"
)

// User representa um usuário do sistema
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"not null" json:"name"`
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	PartnerName  string `json:"partner_name"`
}

func (u *User) isValid(step string) error {
	u.trimSpaces()

	if err := u.checkBlankFields(); err != nil {
		return err
	}

	if err := u.validateEmail(); err != nil {
		return err
	}

	if err := u.validatePassword(); err != nil {
		return err
	}

	if err := u.hashPasswordIfNeeded(step); err != nil {
		return err
	}

	return nil
}

func (u *User) checkBlankFields() error {
		switch {
	case u.Name == "":
		return errors.New("name cannot be empty")
	case u.Email == "":
		return errors.New("email cannot be empty")
	case u.PartnerName == "":
		return errors.New("partner name cannot be empty")
	case u.PasswordHash == "":
		return errors.New("password hash cannot be empty")
	}

	return nil
}

func (u *User) trimSpaces() {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	u.PartnerName = strings.TrimSpace(u.PartnerName)
}

func (u *User) validateEmail() error {
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return errors.New("invalid email format")
	}
	return nil
}

func (u *User) validatePassword() error {
	pass := u.PasswordHash

	if len(pass) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if match, err := regexp.MatchString(`\d`, pass); err != nil || !match {
		return errors.New("password must contain at least one number")
	}
	if match, err := regexp.MatchString(`[A-Z]`, pass); err != nil || !match {
		return errors.New("password must contain at least one uppercase letter")
	}
	if match, err := regexp.MatchString(`[a-z]`, pass); err != nil || !match {
		return errors.New("password must contain at least one lowercase letter")
	}
	if match, err := regexp.MatchString(`[!@#$%^&*(),.?":{}|<>]`, pass); err != nil || !match {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (u *User) hashPasswordIfNeeded(step string) error {
	if step == "register" {
		hashed, err := security.EncryptPassword(u.PasswordHash)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashed)
	}
	return nil
}