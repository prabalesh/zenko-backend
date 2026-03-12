package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prabalesh/zenko-backend/internal/api"
	"github.com/prabalesh/zenko-backend/internal/api/handlers"
	"github.com/prabalesh/zenko-backend/internal/config"
	sqlc "github.com/prabalesh/zenko-backend/internal/db/sqlc"
	"github.com/prabalesh/zenko-backend/internal/services/auth"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Set logging level based on environment
	if cfg.Env == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctx := context.Background()

	// Initialize Database Pool
	dbPool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("database ping failed")
	}

	queries := sqlc.New(dbPool)

	// Initialize Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse redis url")
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	// Initialize Services
	jwtService, err := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry, redisClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize jwt service")
	}
	oauthService := auth.NewOAuthService(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL, redisClient)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(oauthService, jwtService, queries)

	// Setup router
	server := api.SetupRouter(cfg, authHandler, jwtService)

	// Start server
	log.Info().Str("port", cfg.ServerPort).Msg("starting server")
	if err := server.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatal().Err(err).Msg("server startup failed")
	}
}
