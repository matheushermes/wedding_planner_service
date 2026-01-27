package models

import (
	"time"

	"gorm.io/gorm"
)

// Budget representa o orçamento do casamento
type Budget struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	WeddingID    uint    `gorm:"not null;uniqueIndex" json:"wedding_id"`
	Wedding      Wedding `gorm:"foreignKey:WeddingID" json:"-"`
	TotalBudget  float64 `gorm:"not null" json:"total_budget"`
	TotalSpent   float64 `gorm:"default:0" json:"total_spent"`
	TotalPlanned float64 `gorm:"default:0" json:"total_planned"`
}

// Expense representa um gasto
type Expense struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	WeddingID   uint            `gorm:"not null" json:"wedding_id"`
	Wedding     Wedding         `gorm:"foreignKey:WeddingID" json:"-"`
	Category    ExpenseCategory `gorm:"type:varchar(50);not null" json:"category"`
	Description string          `gorm:"type:text" json:"description"`
	Amount      float64         `gorm:"not null" json:"amount"`
	Status      ExpenseStatus   `gorm:"type:varchar(20);default:'planned'" json:"status"`
}

// ExpenseCategory representa as categorias de gastos
type ExpenseCategory string

const (
	ExpenseCategoryFood        ExpenseCategory = "food"
	ExpenseCategoryDecoration  ExpenseCategory = "decoration"
	ExpenseCategoryClothing    ExpenseCategory = "clothing"
	ExpenseCategoryPhotography ExpenseCategory = "photography"
	ExpenseCategoryMusic       ExpenseCategory = "music"
	ExpenseCategoryVenue       ExpenseCategory = "venue"
	ExpenseCategoryOther       ExpenseCategory = "other"
)

// ExpenseStatus representa o status do gasto
type ExpenseStatus string

const (
	ExpenseStatusPlanned ExpenseStatus = "planned"
	ExpenseStatusPaid    ExpenseStatus = "paid"
)
