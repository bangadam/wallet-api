package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func NewPostgresConnection(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Host, config.User, config.Password, config.Name, config.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
