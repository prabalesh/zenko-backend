package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JWTService interface {
	GenerateTokens(userID, username string, elo int) (string, string, error)
	ValidateAccessToken(tokenStr string) (*jwt.MapClaims, error)
	ValidateRefreshToken(ctx context.Context, tokenStr string) (string, string, error)
	InvalidateRefreshToken(ctx context.Context, userID, tokenID string) error
}

type jwtService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	redisClient   *redis.Client
}

func NewJWTService(secret string, accessExp, refreshExp string, redisClient *redis.Client) (JWTService, error) {
	accessDuration, err := time.ParseDuration(accessExp)
	if err != nil {
		return nil, fmt.Errorf("invalid access expiry format: %w", err)
	}
	refreshDuration, err := time.ParseDuration(refreshExp)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh expiry format: %w", err)
	}

	return &jwtService{
		secretKey:     []byte(secret),
		accessExpiry:  accessDuration,
		refreshExpiry: refreshDuration,
		redisClient:   redisClient,
	}, nil
}

func (s *jwtService) GenerateTokens(userID, username string, elo int) (string, string, error) {
	// Access Token
	accessClaims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"elo":      elo,
		"exp":      time.Now().Add(s.accessExpiry).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	tokenID := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"sub":  userID,
		"jti":  tokenID,
		"exp":  time.Now().Add(s.refreshExpiry).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token ID in Redis to allow blacklisting/revocation
	redisKey := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	err = s.redisClient.Set(context.Background(), redisKey, true, s.refreshExpiry).Err()
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token in redis: %w", err)
	}

	return accessStr, refreshStr, nil
}

func (s *jwtService) ValidateAccessToken(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims["type"] != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return &claims, nil
}

func (s *jwtService) ValidateRefreshToken(ctx context.Context, tokenStr string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}

	if claims["type"] != "refresh" {
		return "", "", fmt.Errorf("invalid token type")
	}

	userID := claims["sub"].(string)
	tokenID := claims["jti"].(string)

	redisKey := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	exists, err := s.redisClient.Exists(ctx, redisKey).Result()
	if err != nil {
		return "", "", fmt.Errorf("failed to check token in redis: %w", err)
	}
	if exists == 0 {
		return "", "", fmt.Errorf("refresh token revoked or expired")
	}

	return userID, tokenID, nil
}

func (s *jwtService) ValidateRefreshTokenDirectly(ctx context.Context, userID, tokenID string) error {
	redisKey := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	exists, err := s.redisClient.Exists(ctx, redisKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check token in redis: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("refresh token revoked or expired")
	}
	return nil
}

func (s *jwtService) InvalidateRefreshToken(ctx context.Context, userID, tokenID string) error {
	redisKey := fmt.Sprintf("refresh:%s:%s", userID, tokenID)
	return s.redisClient.Del(ctx, redisKey).Err()
}
