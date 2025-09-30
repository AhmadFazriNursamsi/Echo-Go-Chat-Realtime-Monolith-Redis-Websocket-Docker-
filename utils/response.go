package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ApiResponse struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

// Success response
func Success(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, ApiResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error response
func Error(c echo.Context, code int, message interface{}) error {
	return c.JSON(code, ApiResponse{
		Status:  "error",
		Message: message,
		Data:    nil,
	})
}
