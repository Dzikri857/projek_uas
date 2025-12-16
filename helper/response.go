package helper

import (
	"projek_uas/app/model"

	"github.com/gofiber/fiber/v2"
)

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(model.Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.Response{
		Status:  "error",
		Message: message,
	})
}

func PaginatedResponse(c *fiber.Ctx, data interface{}, pagination model.Pagination) error {
	return c.JSON(model.PaginatedResponse{
		Status: "success",
		Data:   data,
		Meta:   pagination,
	})
}
