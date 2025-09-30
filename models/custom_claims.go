package models

import "github.com/golang-jwt/jwt/v5"

type CustomClaims struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Roleid   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
	RoomsId  []uint `json:"RoomsId"` // ✅ harus sama persis

	jwt.RegisteredClaims
}
