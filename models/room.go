package models

import "gorm.io/gorm"

type Rooms struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex"`
	IsGroupChat bool   `json:"is_group_chat"`
	Users       []User `gorm:"many2many:room_members;"`
}
