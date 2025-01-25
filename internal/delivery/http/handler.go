package http

import (
	"net/http"

	"github.com/bangadam/wallet-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userUsecase        *usecase.UserUsecase
	transactionUsecase *usecase.TransactionUsecase
}

func NewHandler(userUsecase *usecase.UserUsecase, transactionUsecase *usecase.TransactionUsecase) *Handler {
	return &Handler{
		userUsecase:        userUsecase,
		transactionUsecase: transactionUsecase,
	}
}

type RegisterRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Pin         string `json:"pin" binding:"required,len=6"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Pin         string `json:"pin" binding:"required"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remarks string  `json:"remarks" binding:"required"`
}

type TransferRequest struct {
	TargetUser string  `json:"target_user" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Remarks    string  `json:"remarks" binding:"required"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Address   string `json:"address" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUsecase.Register(req.FirstName, req.LastName, req.PhoneNumber, req.Address, req.Pin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": user,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userUsecase.Login(req.PhoneNumber, req.Pin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

func (h *Handler) TopUp(c *gin.Context) {
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	tx, err := h.transactionUsecase.TopUp(userID.(uuid.UUID), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": tx,
	})
}

func (h *Handler) Payment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	tx, err := h.transactionUsecase.Payment(userID.(uuid.UUID), req.Amount, req.Remarks)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": tx,
	})
}

func (h *Handler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID, err := uuid.Parse(req.TargetUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target user ID"})
		return
	}

	userID, _ := c.Get("user_id")
	tx, err := h.transactionUsecase.Transfer(userID.(uuid.UUID), targetUserID, req.Amount, req.Remarks)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": tx,
	})
}

func (h *Handler) GetTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	transactions, err := h.transactionUsecase.GetTransactionsByUserID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": transactions,
	})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	user, err := h.userUsecase.UpdateProfile(userID.(uuid.UUID), req.FirstName, req.LastName, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "SUCCESS",
		"result": user,
	})
}
