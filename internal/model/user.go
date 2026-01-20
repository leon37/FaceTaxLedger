package model

import "time"

type User struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username  string    `gorm:"type:varchar(100);not null;unique" json:"username"`
	Email     string    `gorm:"type:varchar(255);not null;unique;index" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthClaims struct {
	UserID string `json:"user_id"`
}
