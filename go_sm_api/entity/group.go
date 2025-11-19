package entity

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name    string `json:"name" gorm:"uniqueIndex;size:191"`
	OwnerID string `json:"owner_id" gorm:"index;size:64"`
}

type GroupMember struct {
	gorm.Model
	GroupID uint   `json:"group_id" gorm:"index"`
	UserID  string `json:"user_id" gorm:"index;size:64"`
}
