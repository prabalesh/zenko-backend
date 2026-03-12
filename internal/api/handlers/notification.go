package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
	"github.com/prabalesh/zenko-backend/internal/services/notification"
)

type NotificationHandler struct {
	notificationService notification.NotificationService
	validate            *validator.Validate
}

func NewNotificationHandler(ns notification.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: ns,
		validate:            validator.New(),
	}
}

func (h *NotificationHandler) GetPreferences(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.notificationService.GetPreferences(ctx, userID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *NotificationHandler) UpdatePreferences(c *fiber.Ctx) error {
	var req dto.UpdateNotifPrefsReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.notificationService.UpdatePreferences(ctx, userID, &req); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "preferences updated successfully"})
}

func (h *NotificationHandler) RegisterFCMToken(c *fiber.Ctx) error {
	var req dto.RegisterFCMReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.notificationService.RegisterFCMToken(ctx, userID, &req); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "FCM token registered"})
}

func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	cursorP := c.Query("cursor", "")
	var cursor *string
	if cursorP != "" {
		cursor = &cursorP
	}

	limit := 50
	if limitQ := c.Query("limit"); limitQ != "" {
		if l, err := strconv.Atoi(limitQ); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.notificationService.GetNotifications(ctx, userID, cursor, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	notifID := c.Params("id")
	if notifID == "" {
		return errors.BadRequest("notification id required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.notificationService.MarkAsRead(ctx, userID, notifID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.notificationService.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "all marked as read"})
}
