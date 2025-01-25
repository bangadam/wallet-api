package repository

import (
	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *domain.Transaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	return r.db.Create(tx).Error
}

func (r *transactionRepository) GetByUserID(userID uuid.UUID) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByID(id uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(tx *domain.Transaction) error {
	return r.db.Save(tx).Error
}
