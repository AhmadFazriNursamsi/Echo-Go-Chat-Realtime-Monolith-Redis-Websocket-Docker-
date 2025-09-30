package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Messages struct {
	gorm.Model
	RoomId   uint   `json:"room_id"`
	SenderId uint   `json:"sender_id"`
	Content  string `json:"content"`
	MsgType  string `gorm:"column:type" json:"type"` // mapping ke kolom `type`
	Status   string `json:"status"`

	// ✅ Read receipts
	ReadBy      pq.Int64Array `gorm:"type:integer[]"`
	DeliveredTo pq.Int64Array `gorm:"type:integer[]"`
}
