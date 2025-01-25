package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

const (
	TaskTransfer = "task:transfer"
)

type Config struct {
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	DashboardPort int
}

type TransferPayload struct {
	TransactionID string  `json:"transaction_id"`
	FromUserID    string  `json:"from_user_id"`
	ToUserID      string  `json:"to_user_id"`
	Amount        float64 `json:"amount"`
}

type QueueService struct {
	client        *asynq.Client
	server        *asynq.Server
	monitor       *asynqmon.HTTPHandler
	dashboardPort int
}

func NewQueueService(config *Config) *QueueService {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	}

	client := asynq.NewClient(redisOpt)
	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
	})

	// Create Asynq monitor HTTP handler
	monitor := asynqmon.New(asynqmon.Options{
		RootPath:     "/monitoring", // RootPath specifies the root for asynqmon app
		RedisConnOpt: redisOpt,
	})

	return &QueueService{
		client:        client,
		server:        server,
		monitor:       monitor,
		dashboardPort: config.DashboardPort,
	}
}

func (s *QueueService) EnqueueTransfer(payload *TransferPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal transfer payload: %w", err)
	}

	task := asynq.NewTask(TaskTransfer, jsonPayload)
	_, err = s.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue transfer task: %w", err)
	}

	return nil
}

func (s *QueueService) Start(handler func(task *asynq.Task) error) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTransfer, func(ctx context.Context, task *asynq.Task) error {
		return handler(task)
	})

	// Start the monitoring dashboard in a separate goroutine
	go func() {
		monitorMux := http.NewServeMux()
		monitorMux.Handle("/monitoring/", s.monitor)
		monitorMux.Handle("/monitoring", http.RedirectHandler("/monitoring/", http.StatusPermanentRedirect))

		addr := fmt.Sprintf(":%d", s.dashboardPort)
		if err := http.ListenAndServe(addr, monitorMux); err != nil {
			fmt.Printf("Failed to start monitoring dashboard: %v\n", err)
		}
	}()

	if err := s.server.Start(mux); err != nil {
		return fmt.Errorf("failed to start queue server: %w", err)
	}

	return nil
}

func (s *QueueService) Stop() {
	s.server.Stop()
	s.client.Close()
}
