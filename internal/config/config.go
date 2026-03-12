package config

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort         string `mapstructure:"SERVER_PORT" validate:"required"`
	Env                string `mapstructure:"ENV" validate:"required,oneof=development staging production"`
	DBURL              string `mapstructure:"DB_URL" validate:"required,url"`
	RedisURL           string `mapstructure:"REDIS_URL" validate:"required"`
	JWTSecret          string `mapstructure:"JWT_SECRET" validate:"required"`
	JWTAccessExpiry    string `mapstructure:"JWT_ACCESS_EXPIRY" validate:"required"`
	JWTRefreshExpiry   string `mapstructure:"JWT_REFRESH_EXPIRY" validate:"required"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`
	FCMServerKey       string `mapstructure:"FCM_SERVER_KEY"`
	AppName            string `mapstructure:"APP_NAME" validate:"required"`
	AllowedOrigins     string `mapstructure:"ALLOWED_ORIGINS" validate:"required"`
}

var validate = validator.New()

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}
