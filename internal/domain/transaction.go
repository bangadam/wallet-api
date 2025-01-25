package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeCredit TransactionType = "CREDIT"
	TransactionTypeDebit  TransactionType = "DEBIT"

	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusPending TransactionStatus = "PENDING"
)

type Transaction struct {
	ID            uuid.UUID         `gorm:"type:uuid;primary_key" json:"transaction_id"`
	UserID        uuid.UUID         `json:"user_id"`
	Type          TransactionType   `json:"transaction_type"`
	Status        TransactionStatus `json:"status"`
	Amount        float64           `json:"amount"`
	Remarks       string            `json:"remarks"`
	BalanceBefore float64           `json:"balance_before"`
	BalanceAfter  float64           `json:"balance_after"`
	ReferenceID   uuid.UUID         `json:"reference_id,omitempty"`
	ReferenceType string            `json:"reference_type,omitempty"`
	TargetUserID  *uuid.UUID        `json:"target_user_id,omitempty"`
	CreatedAt     time.Time         `json:"created_date"`
	UpdatedAt     time.Time         `json:"updated_date"`
}

type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByUserID(userID uuid.UUID) ([]Transaction, error)
	GetByID(id uuid.UUID) (*Transaction, error)
	Update(tx *Transaction) error
}
