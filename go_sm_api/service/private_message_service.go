package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/abeme/go_sm_api/entity"
)

// PrivateMessageService defines operations for direct messages.
type PrivateMessageService interface {
	Send(senderID, recipientID, body string) (*entity.PrivateMessage, error)
	ListConversation(userID, otherUserID string, limit int, beforeID uint) ([]entity.PrivateMessage, error)
	MarkRead(recipientID, senderID string, ids []uint) (int64, error)
}

type DBPrivateMessageService struct {
	db *gorm.DB
}

func NewPrivateMessageService(db *gorm.DB) *DBPrivateMessageService {
	return &DBPrivateMessageService{db: db}
}

func (s *DBPrivateMessageService) Send(senderID, recipientID, body string) (*entity.PrivateMessage, error) {
	if senderID == recipientID {
		return nil, errors.New("cannot send to self")
	}
	pm := &entity.PrivateMessage{SenderID: senderID, RecipientID: recipientID, Body: body}
	if err := s.db.Create(pm).Error; err != nil {
		return nil, err
	}
	return pm, nil
}

// ListConversation returns messages between two users ordered newest first.
// If beforeID > 0, returns messages with ID < beforeID for pagination.
func (s *DBPrivateMessageService) ListConversation(userID, otherUserID string, limit int, beforeID uint) ([]entity.PrivateMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var msgs []entity.PrivateMessage
	q := s.db.Model(&entity.PrivateMessage{}).
		Where("((sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?))", userID, otherUserID, otherUserID, userID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	if err := q.Order("id DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

// MarkRead sets ReadAt for specified message IDs where recipient is recipientID and sender is senderID.
func (s *DBPrivateMessageService) MarkRead(recipientID, senderID string, ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	now := time.Now()
	res := s.db.Model(&entity.PrivateMessage{}).
		Where("recipient_id = ? AND sender_id = ? AND id IN (?) AND read_at IS NULL", recipientID, senderID, ids).
		Updates(map[string]interface{}{"read_at": &now})
	return res.RowsAffected, res.Error
}
