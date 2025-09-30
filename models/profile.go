package models

import "gorm.io/gorm"

type Profile struct {
	gorm.Model
	UserID   uint `gorm:"uniqueIndex"`
	FullName string
	Phone    string `gorm:"uniqueIndex"`
	Address  string
}
