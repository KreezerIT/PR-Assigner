package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config общая конфигурация приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
}

// ServerConfig настройки HTTP‑сервера
type ServerConfig struct {
	Port string
}

// DatabaseConfig настройки подключения к БД и параметры пула соединений
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxRetries      int
	RetryDelay      time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// LoggerConfig настройки логгера
type LoggerConfig struct {
	Level string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "prservice"),
			Password:        getEnv("DB_PASSWORD", "prservice_pass"),
			DBName:          getEnv("DB_NAME", "pr_reviewer"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxRetries:      getEnvAsInt("DB_MAX_RETRIES", 5),
			RetryDelay:      getEnvAsDuration("DB_RETRY_DELAY", 2*time.Second),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

// DSN формирует строку подключения к БД
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	dur, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultValue
	}
	return dur
}
