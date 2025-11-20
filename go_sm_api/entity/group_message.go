package entity

import "time"

type GroupMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	GroupID   uint      `json:"group_id" gorm:"index"`
	SenderID  string    `json:"sender_id" gorm:"index;size:64"`
	Body      string    `json:"body" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}
