package repository

import "time"

type Token struct {
	Id        int       `gorm:"primaryKey;autoIncrement"`
	Token     string    `gorm:"size:255;not null;uniqueIndex"`
	UserId    int       `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
