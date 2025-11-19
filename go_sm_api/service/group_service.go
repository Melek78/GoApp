package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/abeme/go_sm_api/entity"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrGroupExists = errors.New("group already exists")
)

type GroupService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewGroupService(db *gorm.DB, rdb *redis.Client) *GroupService {
	return &GroupService{db: db, rdb: rdb}
}

func (s *GroupService) CreateGroup(name string, ownerID string) (*entity.Group, error) {
	g := &entity.Group{Name: name, OwnerID: ownerID}
	if err := s.db.Create(g).Error; err != nil {
		return nil, err
	}
	// add owner as member (user IDs are strings)
	gm := &entity.GroupMember{GroupID: g.ID, UserID: ownerID}
	if err := s.db.Create(gm).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GroupService) JoinGroup(groupID uint, userID string) error {
	// check exists
	var count int64
	if err := s.db.Model(&entity.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // already member
	}
	gm := &entity.GroupMember{GroupID: groupID, UserID: userID}
	if err := s.db.Create(gm).Error; err != nil {
		return err
	}
	return nil
}

func (s *GroupService) GetMembers(groupID uint) ([]string, error) {
	var members []entity.GroupMember
	if err := s.db.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids, nil
}

func (s *GroupService) PublishGroupMessage(ctx context.Context, groupID uint, msg string) error {
	ch := "group:" + strconv.FormatUint(uint64(groupID), 10)
	return s.rdb.Publish(ctx, ch, msg).Err()
}
