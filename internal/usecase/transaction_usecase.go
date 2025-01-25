package usecase

import (
	"errors"
	"time"

	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/bangadam/wallet-api/pkg/queue"
	"github.com/google/uuid"
)

type TransactionUsecase struct {
	transactionRepo domain.TransactionRepository
	userRepo        domain.UserRepository
	queueService    *queue.QueueService
}

func NewTransactionUsecase(
	transactionRepo domain.TransactionRepository,
	userRepo domain.UserRepository,
	queueService *queue.QueueService,
) *TransactionUsecase {
	return &TransactionUsecase{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		queueService:    queueService,
	}
}

func (u *TransactionUsecase) TopUp(userID uuid.UUID, amount float64) (*domain.Transaction, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	tx := &domain.Transaction{
		ID:            uuid.New(),
		UserID:        userID,
		Type:          domain.TransactionTypeCredit,
		Status:        domain.TransactionStatusSuccess,
		Amount:        amount,
		BalanceBefore: user.Balance,
		BalanceAfter:  user.Balance + amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	if err := u.userRepo.UpdateBalance(userID, tx.BalanceAfter); err != nil {
		return nil, err
	}

	return tx, nil
}

func (u *TransactionUsecase) Payment(userID uuid.UUID, amount float64, remarks string) (*domain.Transaction, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Balance < amount {
		return nil, errors.New("balance is not enough")
	}

	tx := &domain.Transaction{
		ID:            uuid.New(),
		UserID:        userID,
		Type:          domain.TransactionTypeDebit,
		Status:        domain.TransactionStatusSuccess,
		Amount:        amount,
		Remarks:       remarks,
		BalanceBefore: user.Balance,
		BalanceAfter:  user.Balance - amount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	if err := u.userRepo.UpdateBalance(userID, tx.BalanceAfter); err != nil {
		return nil, err
	}

	return tx, nil
}

func (u *TransactionUsecase) Transfer(fromUserID, toUserID uuid.UUID, amount float64, remarks string) (*domain.Transaction, error) {
	fromUser, err := u.userRepo.GetByID(fromUserID)
	if err != nil {
		return nil, err
	}

	if fromUser.Balance < amount {
		return nil, errors.New("balance is not enough")
	}

	// Create pending transaction
	tx := &domain.Transaction{
		ID:            uuid.New(),
		UserID:        fromUserID,
		Type:          domain.TransactionTypeDebit,
		Status:        domain.TransactionStatusPending,
		Amount:        amount,
		Remarks:       remarks,
		BalanceBefore: fromUser.Balance,
		BalanceAfter:  fromUser.Balance - amount,
		TargetUserID:  &toUserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	// Enqueue transfer task
	err = u.queueService.EnqueueTransfer(&queue.TransferPayload{
		TransactionID: tx.ID.String(),
		FromUserID:    fromUserID.String(),
		ToUserID:      toUserID.String(),
		Amount:        amount,
	})
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (u *TransactionUsecase) GetTransactionsByUserID(userID uuid.UUID) ([]domain.Transaction, error) {
	return u.transactionRepo.GetByUserID(userID)
}

func (u *TransactionUsecase) ProcessTransfer(transactionID, fromUserID, toUserID string, amount float64) error {
	txID, err := uuid.Parse(transactionID)
	if err != nil {
		return err
	}

	fromUID, err := uuid.Parse(fromUserID)
	if err != nil {
		return err
	}

	toUID, err := uuid.Parse(toUserID)
	if err != nil {
		return err
	}

	// Get transaction
	tx, err := u.transactionRepo.GetByID(txID)
	if err != nil {
		return err
	}

	// Update sender's balance
	if err := u.userRepo.UpdateBalance(fromUID, tx.BalanceAfter); err != nil {
		tx.Status = domain.TransactionStatusFailed
		u.transactionRepo.Update(tx)
		return err
	}

	// Create recipient's transaction
	recipientTx := &domain.Transaction{
		ID:            uuid.New(),
		UserID:        toUID,
		Type:          domain.TransactionTypeCredit,
		Status:        domain.TransactionStatusSuccess,
		Amount:        amount,
		Remarks:       tx.Remarks,
		ReferenceID:   tx.ID,
		ReferenceType: "transfer",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	toUser, err := u.userRepo.GetByID(toUID)
	if err != nil {
		return err
	}

	recipientTx.BalanceBefore = toUser.Balance
	recipientTx.BalanceAfter = toUser.Balance + amount

	if err := u.transactionRepo.Create(recipientTx); err != nil {
		return err
	}

	// Update recipient's balance
	if err := u.userRepo.UpdateBalance(toUID, recipientTx.BalanceAfter); err != nil {
		return err
	}

	// Update original transaction status
	tx.Status = domain.TransactionStatusSuccess
	return u.transactionRepo.Update(tx)
}
