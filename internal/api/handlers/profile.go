package handlers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
	"github.com/prabalesh/zenko-backend/internal/services/profile"
)

type ProfileHandler struct {
	profileService profile.ProfileService
	validate       *validator.Validate
}

func NewProfileHandler(profileService profile.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		validate:       validator.New(),
	}
}

func (h *ProfileHandler) SetUsername(c *fiber.Ctx) error {
	var req dto.SetUsernameReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.profileService.SetUsername(ctx, userID, req.Username); err != nil {
		return err // Assuming service returns *errors.AppError
	}

	return c.JSON(fiber.Map{"message": "username set successfully"})
}

func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	profileMap, err := h.profileService.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(profileMap)
}

func (h *ProfileHandler) GetPublicProfile(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return errors.BadRequest("username is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	profileMap, err := h.profileService.GetPublicProfile(ctx, username)
	if err != nil {
		return err
	}

	return c.JSON(profileMap)
}

func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	var req dto.UpdateProfileReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.profileService.UpdateProfile(ctx, userID, &req); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "profile updated successfully"})
}

func (h *ProfileHandler) ChangeUsername(c *fiber.Ctx) error {
	var req dto.ChangeUsernameReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.profileService.ChangeUsername(ctx, userID, req.Username); err != nil {
		// Ensure to map 429 if returning it directly, our AppError defaults to general statuses unless explicitly typed
		return err
	}

	return c.JSON(fiber.Map{"message": "username changed successfully"})
}

func (h *ProfileHandler) UpdateSocialLinks(c *fiber.Ctx) error {
	var req dto.UpdateSocialLinksReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.profileService.UpdateSocialLinks(ctx, userID, &req); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "social links updated successfully"})
}
