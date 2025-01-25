package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bangadam/wallet-api/internal/delivery/http"
	"github.com/bangadam/wallet-api/internal/domain"
	"github.com/bangadam/wallet-api/internal/middleware"
	"github.com/bangadam/wallet-api/internal/repository"
	"github.com/bangadam/wallet-api/internal/usecase"
	"github.com/bangadam/wallet-api/pkg/auth"
	"github.com/bangadam/wallet-api/pkg/database"
	"github.com/bangadam/wallet-api/pkg/queue"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// Setup database connection
	dbConfig := &database.Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		Name:     viper.GetString("database.name"),
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err)
	}

	// Auto migrate database
	err = db.AutoMigrate(&domain.User{}, &domain.Transaction{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
	}

	// Setup JWT service
	jwtConfig := &auth.JWTConfig{
		Secret:            viper.GetString("jwt.secret"),
		ExpirationHours:   viper.GetInt("jwt.expiration"),
		RefreshExpiration: viper.GetInt("jwt.refresh_expiration"),
	}
	jwtService := auth.NewJWTService(jwtConfig)

	// Setup queue service
	queueConfig := &queue.Config{
		RedisHost:     viper.GetString("redis.host"),
		RedisPort:     viper.GetInt("redis.port"),
		RedisPassword: viper.GetString("redis.password"),
		RedisDB:       viper.GetInt("redis.db"),
		DashboardPort: viper.GetInt("redis.dashboard_port"),
	}
	queueService := queue.NewQueueService(queueConfig)

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Setup usecases
	userUsecase := usecase.NewUserUsecase(userRepo, jwtService)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, userRepo, queueService)

	// Setup HTTP handler
	handler := http.NewHandler(userUsecase, transactionUsecase)

	// Setup background worker for transfer processing
	go func() {
		err := queueService.Start(func(task *asynq.Task) error {
			var payload queue.TransferPayload
			if err := json.Unmarshal(task.Payload(), &payload); err != nil {
				return err
			}

			return transactionUsecase.ProcessTransfer(
				payload.TransactionID,
				payload.FromUserID,
				payload.ToUserID,
				payload.Amount,
			)
		})
		if err != nil {
			log.Printf("Failed to start queue worker: %s", err)
		}
	}()

	// Setup Gin router
	router := gin.Default()

	// Public routes
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	// Protected routes
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))
	{
		protected.POST("/topup", handler.TopUp)
		protected.POST("/pay", handler.Payment)
		protected.POST("/transfer", handler.Transfer)
		protected.GET("/transactions", handler.GetTransactions)
		protected.PUT("/profile", handler.UpdateProfile)
	}

	// Start server
	port := viper.GetString("app.port")
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
