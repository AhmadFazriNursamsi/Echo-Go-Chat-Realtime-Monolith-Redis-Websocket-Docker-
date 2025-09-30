package handlers

import (
	"echo-app/database"
	"echo-app/dto"
	"echo-app/models"
	"echo-app/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func CreateProfile(c echo.Context) error {
	// var profile models.Profile

	req := new(dto.ProfileDto)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"errors": utils.FormatValidationError(err)})
	}

	profile := models.Profile{
		UserID:   req.UserID,
		FullName: req.FullName,
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := database.DB.Create(&profile).Error; err != nil {
		if strings.Contains(err.Error(), "user_id") {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "UserId tersebut sudah terdaftar"})
		}
		if strings.Contains(err.Error(), "phone") {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "No Phone tersebut sudah terdaftar"})
		}
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Gagal membuat profile"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "User registered successfully",
		"profile": profile,
	})

}

func GetProfiles(c echo.Context) error {
	var profiles []models.Profile
	database.DB.Preload("User").Find(&profiles)
	return c.JSON(http.StatusOK, profiles)
}
func GetProfileByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var profile models.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Profile not found",
		})
	}

	// Ambil user berdasarkan profile.UserID
	var user models.User
	if err := database.DB.First(&user, profile.UserID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "User not found for this profile",
		})
	}

	// Mapping ke DTO
	resp := dto.ProfileDto{
		UserID:   profile.UserID,
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Address:  profile.Address,
	}
	resp.User.ID = user.ID
	resp.User.Email = user.Email

	return c.JSON(http.StatusOK, resp)
}
func UpdateProfile(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var profile models.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Profile not found"})
	}

	// DTO request
	var req dto.UpdateProfileDto
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	// Buat map updates hanya dari field yang tidak kosong
	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	// Update ke DB
	if len(updates) > 0 {
		if err := database.DB.Model(&profile).Updates(updates).Error; err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
	}

	// Ambil ulang data
	database.DB.First(&profile, id)

	return c.JSON(http.StatusOK, profile)
}

func GetMyProfile(c echo.Context) error {
	// Ambil token dari context (sudah di-inject Echo JWT middleware)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)

	// Ambil user_id dari claims
	userID := uint(claims["user_id"].(float64))

	// Cari user lengkap dengan profile + role
	var user models.User
	if err := database.DB.Preload("Profile").Preload("Role").First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	// Response rapi
	return c.JSON(http.StatusOK, echo.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role.Name,
		"profile": echo.Map{
			"full_name": user.Profile.FullName,
			"phone":     user.Profile.Phone,
			"address":   user.Profile.Address,
		},
	})
}

func DeleteProfiles(c echo.Context) error {
	id := c.Param("id")

	var profile models.Profile
	if err := database.DB.Preload("User").Delete(&profile, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "Profile not found",
		})
	}

	return c.JSON(http.StatusOK, profile)
}
