package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email             string `gorm:"uniqueIndex"`
	Password          string
	RoleID            uint
	Role              Role
	Profile           Profile
	PasswordChangedAt time.Time `json:"-"` // tidak dikirim ke response
	Rooms             []Rooms   `gorm:"many2many:room_members;"`
}
