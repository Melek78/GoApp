package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/abeme/go_sm_api/entity"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
	ErrUserNotFound = errors.New("user not found")
)

// UserService interface abstracts user ops
type UserService interface {
	CreateUser(email, password string) (*entity.User, error)
	Authenticate(email, password string) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	GetByID(id string) (*entity.User, error)
}

type DBUserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *DBUserService {
	return &DBUserService{db: db}
}

func generateID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *DBUserService) CreateUser(email, password string) (*entity.User, error) {
	var cnt int64
	if err := s.db.Model(&entity.User{}).Where("email = ?", email).Count(&cnt).Error; err != nil {
		return nil, err
	}
	if cnt > 0 {
		return nil, ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &entity.User{
		ID:           generateID(8),
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (s *DBUserService) Authenticate(email, password string) (*entity.User, error) {
	var u entity.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCreds
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}
	return &u, nil
}

func (s *DBUserService) GetByEmail(email string) (*entity.User, error) {
	var u entity.User
	if err := s.db.Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *DBUserService) GetByID(id string) (*entity.User, error) {
	var u entity.User
	if err := s.db.Where("id = ?", id).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// ...existing code...
