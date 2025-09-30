package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"echo-app/database"
	"echo-app/dto"
	"echo-app/models"
	"echo-app/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func hashPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashed)
}

// Register
func Register(c echo.Context) error {
	req := new(dto.RegisterUserDto)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err))

	}

	user := models.User{
		Email:    req.Email,
		Password: hashPassword(req.Password),
		RoleID:   req.RoleID,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		// return c.JSON(http.StatusBadRequest, echo.Map{"error": "Email already registered"})
		return utils.Error(c, http.StatusBadRequest, "Email already registered")

	}
	return utils.Success(c, "User registered successfully", user)

}

// Login
func Login(c echo.Context) error {
	req := new(dto.LoginDto)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err)[0])

	}

	var user models.User

	if err := database.DB.Preload("Role").
		Preload("Rooms").
		Preload("Profile").
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {
		return utils.Error(c, http.StatusUnauthorized, "Email tidak terdaftar")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return utils.Error(c, http.StatusUnauthorized, "password salah")

	}

	// kumpulkan room IDs
	var roomIDs []uint
	for _, r := range user.Rooms {
		roomIDs = append(roomIDs, r.ID)
	}

	// siapkan claims
	claims := models.CustomClaims{
		ID:       user.ID,
		Name:     user.Profile.FullName,
		Email:    user.Email,
		Roleid:   user.RoleID,
		RoleName: user.Role.Name,
		RoomsId:  roomIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "Atep Token",
			Subject:   fmt.Sprint(user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(utils.GetJwtSecret())

	if err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Could not generate token")
	}
	return utils.Success(c, "Login success", t)

}

// Forgot Password - generate reset token
func ForgotPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err))
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "Email not registered")
	}

	// generate reset token (JWT berlaku 15 menit)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"reset":   true,
		"iat":     time.Now().Unix(), // issued at
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	resetToken, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	// TODO: kirim via email (production)
	return utils.Success(c, "Reset token generated", echo.Map{
		"reset_token": resetToken, // hanya untuk testing
	})
}

// Reset Password
func ResetPassword(c echo.Context) error {
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err))
	}

	// parse token
	parsed, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsed.Valid {
		return utils.Error(c, http.StatusUnauthorized, "Invalid or expired reset token")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["reset"] != true {
		return utils.Error(c, http.StatusUnauthorized, "Invalid reset token")
	}

	userID := uint(claims["user_id"].(float64))
	tokenIssuedAt := int64(claims["iat"].(float64))

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "User not found")
	}

	// ✅ Pastikan token dibuat setelah password terakhir diganti
	if user.PasswordChangedAt.Unix() > tokenIssuedAt {
		return utils.Error(c, http.StatusUnauthorized, "Token Sudah pernah digunakan")
	}

	// hash new password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)

	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"password":            string(hashed),
		"password_changed_at": time.Now(),
	}).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to reset password")
	}

	return utils.Success(c, "Password has been reset successfully", nil)
}

func ChangePassword(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err))
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return utils.Error(c, http.StatusNotFound, "User not found")
	}

	// cek old password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)) != nil {
		return utils.Error(c, http.StatusUnauthorized, "Old password is incorrect")
	}

	// hash new password
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
	user.Password = string(hashed)
	database.DB.Save(&user)

	return utils.Success(c, "Password changed successfully", nil)
}

func ChangeEmail(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var req struct {
		NewEmail string `json:"new_email" validate:"required,email"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, "Invalid request")
	}
	if err := c.Validate(&req); err != nil {
		return utils.Error(c, http.StatusBadRequest, utils.FormatValidationError(err))
	}

	// cek email unik
	var existing models.User
	if err := database.DB.Where("email = ?", req.NewEmail).First(&existing).Error; err == nil {
		return utils.Error(c, http.StatusBadRequest, "Email already registered")
	}

	// update email
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("email", req.NewEmail).Error; err != nil {
		return utils.Error(c, http.StatusInternalServerError, "Failed to update email")
	}

	return utils.Success(c, "Email changed successfully", echo.Map{"new_email": req.NewEmail})
}
