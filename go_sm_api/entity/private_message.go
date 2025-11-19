package entity

import "time"

// PrivateMessage represents a direct message between two users.
// SenderID and RecipientID reference User.ID (string).
// ReadAt is null until the recipient marks the message as read.
type PrivateMessage struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	SenderID    string     `json:"sender_id" gorm:"index;size:64"`
	RecipientID string     `json:"recipient_id" gorm:"index;size:64"`
	Body        string     `json:"body" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at"`
	ReadAt      *time.Time `json:"read_at"`
}
