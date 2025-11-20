package service

import (
	"github.com/abeme/go_sm_api/entity"
	"gorm.io/gorm"
)

type GroupMessageService interface {
	Send(groupID uint, senderID, body string) (*entity.GroupMessage, error)
	List(groupID uint, limit int, beforeID uint) ([]entity.GroupMessage, error)
}

type DBGroupMessageService struct {
	db *gorm.DB
}

func NewGroupMessageService(db *gorm.DB) *DBGroupMessageService {
	return &DBGroupMessageService{db: db}
}

func (s *DBGroupMessageService) Send(groupID uint, senderID, body string) (*entity.GroupMessage, error) {
	gm := &entity.GroupMessage{GroupID: groupID, SenderID: senderID, Body: body}
	if err := s.db.Create(gm).Error; err != nil {
		return nil, err
	}
	return gm, nil
}

func (s *DBGroupMessageService) List(groupID uint, limit int, beforeID uint) ([]entity.GroupMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var msgs []entity.GroupMessage
	q := s.db.Model(&entity.GroupMessage{}).Where("group_id = ?", groupID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	if err := q.Order("id DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}
