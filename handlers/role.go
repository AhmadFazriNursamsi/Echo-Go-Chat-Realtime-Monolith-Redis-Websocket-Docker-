package handlers

import (
	"echo-app/database"
	"echo-app/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateRole(c echo.Context) error {
	var role models.Role
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
	}

	if err := database.DB.Create(&role).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Role Already Exist"})
	}

	return c.JSON(http.StatusOK, role)
}

func GetRoles(c echo.Context) error {
	var roles []models.Role
	database.DB.Find(&roles)
	return c.JSON(http.StatusOK, roles)
}

func GetRolesByid(c echo.Context) error {
	id := c.Param("id")
	var roles []models.Role
	if err := database.DB.First(&roles, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Role not found",
		})
	}
	return c.JSON(http.StatusOK, roles)
}
