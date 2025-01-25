package usecase

import (
	"errors"
	"time"

	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/bangadam/wallet-api/pkg/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo   domain.UserRepository
	jwtService *auth.JWTService
}

func NewUserUsecase(userRepo domain.UserRepository, jwtService *auth.JWTService) *UserUsecase {
	return &UserUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (u *UserUsecase) Register(firstName, lastName, phoneNumber, address, pin string) (*domain.User, error) {
	// Check if phone number already exists
	existingUser, err := u.userRepo.GetByPhoneNumber(phoneNumber)
	if err == nil && existingUser != nil {
		return nil, errors.New("phone number already registered")
	}

	// Hash PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:          uuid.New(),
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
		Address:     address,
		Pin:         string(hashedPin),
		Balance:     0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) Login(phoneNumber, pin string) (string, string, error) {
	user, err := u.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return "", "", errors.New("phone number and PIN don't match")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Pin), []byte(pin)); err != nil {
		return "", "", errors.New("phone number and PIN don't match")
	}

	accessToken, refreshToken, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *UserUsecase) UpdateProfile(userID uuid.UUID, firstName, lastName, address string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Address = address
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) GetUserByID(userID uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetByID(userID)
}
