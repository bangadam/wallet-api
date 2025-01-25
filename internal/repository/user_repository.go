package repository

import (
	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.Create(user).Error
}

func (r *userRepository) GetByPhoneNumber(phoneNumber string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdateBalance(userID uuid.UUID, amount float64) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("balance", amount).
		Error
}
