package errors

import (
	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    fiber.StatusUnauthorized,
		Message: msg,
	}
}

func BadRequest(msg string) *AppError {
	return &AppError{
		Code:    fiber.StatusBadRequest,
		Message: msg,
	}
}

func Internal(msg string) *AppError {
	return &AppError{
		Code:    fiber.StatusInternalServerError,
		Message: msg,
	}
}
