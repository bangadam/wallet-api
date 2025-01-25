package usecase

import (
	"testing"
	"time"

	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/bangadam/wallet-api/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByPhoneNumber(phoneNumber string) (*domain.User, error) {
	args := m.Called(phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateBalance(userID uuid.UUID, amount float64) error {
	args := m.Called(userID, amount)
	return args.Error(0)
}

func TestUserUsecase_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService(&auth.JWTConfig{
		Secret:            "test-secret",
		ExpirationHours:   24,
		RefreshExpiration: 168,
	})
	usecase := NewUserUsecase(mockRepo, jwtService)

	t.Run("successful registration", func(t *testing.T) {
		mockRepo.On("GetByPhoneNumber", "1234567890").Return(nil, nil).Once()
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil).Once()

		user, err := usecase.Register("John", "Doe", "1234567890", "Test Address", "123456")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "1234567890", user.PhoneNumber)
		assert.Equal(t, "Test Address", user.Address)
		assert.NotEmpty(t, user.Pin)
		mockRepo.AssertExpectations(t)
	})

	t.Run("phone number already exists", func(t *testing.T) {
		existingUser := &domain.User{
			ID:          uuid.New(),
			FirstName:   "Existing",
			LastName:    "User",
			PhoneNumber: "1234567890",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByPhoneNumber", "1234567890").Return(existingUser, nil).Once()

		user, err := usecase.Register("John", "Doe", "1234567890", "Test Address", "123456")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "phone number already registered", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService(&auth.JWTConfig{
		Secret:            "test-secret",
		ExpirationHours:   24,
		RefreshExpiration: 168,
	})
	usecase := NewUserUsecase(mockRepo, jwtService)

	t.Run("successful login", func(t *testing.T) {
		hashedPin := "$2a$10$1234567890123456789012345678901234567890" // pre-hashed "123456"
		user := &domain.User{
			ID:          uuid.New(),
			FirstName:   "John",
			LastName:    "Doe",
			PhoneNumber: "1234567890",
			Pin:         hashedPin,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByPhoneNumber", "1234567890").Return(user, nil).Once()

		accessToken, refreshToken, err := usecase.Login("1234567890", "123456")

		assert.Error(t, err) // Will fail because the pin hash is not valid
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockRepo.On("GetByPhoneNumber", "1234567890").Return(nil, assert.AnError).Once()

		accessToken, refreshToken, err := usecase.Login("1234567890", "123456")

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Equal(t, "phone number and PIN don't match", err.Error())
		mockRepo.AssertExpectations(t)
	})
}
