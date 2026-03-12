package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	sqlc "github.com/prabalesh/zenko-backend/internal/db/sqlc"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
	"github.com/prabalesh/zenko-backend/internal/services/auth"
)

type AuthHandler struct {
	oauthService auth.OAuthService
	jwtService   auth.JWTService
	db           *sqlc.Queries
}

func NewAuthHandler(oauthService auth.OAuthService, jwtService auth.JWTService, db *sqlc.Queries) *AuthHandler {
	return &AuthHandler{
		oauthService: oauthService,
		jwtService:   jwtService,
		db:           db,
	}
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	url, err := h.oauthService.GetAuthURL(ctx)
	if err != nil {
		return errors.Internal("Failed to generate auth url: " + err.Error())
	}

	return c.Redirect(url)
}

func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		return errors.BadRequest("Missing state or code")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.oauthService.VerifyState(ctx, state); err != nil {
		return errors.Unauthorized("Invalid state: " + err.Error())
	}

	googleUser, err := h.oauthService.GetGoogleUserInfo(ctx, code)
	if err != nil {
		return errors.Unauthorized("Failed to get user info: " + err.Error())
	}

	user, err := h.db.GetUserByGoogleID(ctx, googleUser.ID)
	isNewUser := false

	if err != nil {
		if err == pgx.ErrNoRows {
			// Create user
			username := googleUser.Name // Ideally this should be a unique generator or picker
			if username == "" {
				username = "User" + googleUser.ID[:6]
			}
			user, err = h.db.CreateUser(ctx, sqlc.CreateUserParams{
				GoogleID:  googleUser.ID,
				Username:  username,
				AvatarUrl: googleUser.Picture,
			})
			if err != nil {
				return errors.Internal("Failed to create user: " + err.Error())
			}
			isNewUser = true
		} else {
			return errors.Internal("Database error: " + err.Error())
		}
	}

	access, refresh, err := h.jwtService.GenerateTokens(user.ID.String(), user.Username, int(user.Elo))
	if err != nil {
		return errors.Internal("Failed to generate tokens")
	}

	return c.JSON(dto.GoogleCallbackResp{
		AccessToken:  access,
		RefreshToken: refresh,
		IsNewUser:    isNewUser,
		User: dto.UserResp{
			ID:        user.ID.String(),
			Username:  user.Username,
			AvatarURL: user.AvatarUrl,
			Elo:       int(user.Elo),
			IsNewUser: isNewUser,
		},
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	userID, _, err := h.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return errors.Unauthorized(err.Error())
	}

	// Ideally fetch user from DB to get username/elo to pack in the new access token
	// For now, packing empty or default
	access, _, err := h.jwtService.GenerateTokens(userID, "user", 1200) // Needs real user data
	if err != nil {
		return errors.Internal("Failed to generate access token")
	}

	// (Optional) Replace old refresh token with new one - keeping it simple and just issuing access

	return c.JSON(dto.RefreshResp{
		AccessToken: access,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req dto.RefreshReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	userID, tokenID, err := h.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return errors.Unauthorized("Invalid refresh token")
	}

	if err := h.jwtService.InvalidateRefreshToken(ctx, userID, tokenID); err != nil {
		return errors.Internal("Failed to invalidate token")
	}

	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}
