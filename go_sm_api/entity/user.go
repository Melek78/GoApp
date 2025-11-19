package entity

type User struct {
	ID           string `json:"id" gorm:"primaryKey;size:64"`
	Email        string `json:"email" gorm:"uniqueIndex;size:191"`
	PasswordHash string `json:"-" gorm:"size:191"`
}

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
