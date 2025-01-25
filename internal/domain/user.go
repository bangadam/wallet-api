package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"user_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `gorm:"unique" json:"phone_number"`
	Address     string    `json:"address"`
	Pin         string    `json:"-"`
	Balance     float64   `json:"balance"`
	CreatedAt   time.Time `json:"created_date"`
	UpdatedAt   time.Time `json:"updated_date"`
}

type UserRepository interface {
	Create(user *User) error
	GetByPhoneNumber(phoneNumber string) (*User, error)
	GetByID(id uuid.UUID) (*User, error)
	Update(user *User) error
	UpdateBalance(userID uuid.UUID, amount float64) error
}
