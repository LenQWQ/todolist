package repository

import "time"

type User struct {
	Id           int       `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:255;not null;unique"`
	Email        string    `gorm:"size:255;not null;unique"`
	PasswordHash string    `gorm:"size:255;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
