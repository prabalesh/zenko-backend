package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/prabalesh/zenko-backend/internal/pkg/dto"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
	"github.com/prabalesh/zenko-backend/internal/services/friends"
)

type FriendsHandler struct {
	friendsService friends.FriendsService
	validate       *validator.Validate
}

func NewFriendsHandler(friendsService friends.FriendsService) *FriendsHandler {
	return &FriendsHandler{
		friendsService: friendsService,
		validate:       validator.New(),
	}
}

func (h *FriendsHandler) SendRequest(c *fiber.Ctx) error {
	var req dto.SendFriendReqReq
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return errors.BadRequest("Validation failed: " + err.Error())
	}

	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.friendsService.SendFriendRequest(ctx, userID, req.Username); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "friend request sent successfully"})
}

func (h *FriendsHandler) GetFriends(c *fiber.Ctx) error {
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

	resp, err := h.friendsService.GetFriendsList(ctx, userID, cursor, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *FriendsHandler) GetRequests(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.friendsService.GetFriendRequests(ctx, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *FriendsHandler) AcceptRequest(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	friendID := c.Params("id")
	if friendID == "" {
		return errors.BadRequest("friend id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.friendsService.AcceptFriendRequest(ctx, userID, friendID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "friend request accepted"})
}

func (h *FriendsHandler) RemoveFriend(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	friendID := c.Params("id")
	if friendID == "" {
		return errors.BadRequest("friend id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.friendsService.RemoveFriend(ctx, userID, friendID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "friend removed"})
}

func (h *FriendsHandler) BlockUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	friendID := c.Params("id")
	if friendID == "" {
		return errors.BadRequest("user id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.friendsService.BlockUser(ctx, userID, friendID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "user blocked"})
}

func (h *FriendsHandler) UnblockUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	friendID := c.Params("id")
	if friendID == "" {
		return errors.BadRequest("user id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.friendsService.UnblockUser(ctx, userID, friendID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "user unblocked"})
}
