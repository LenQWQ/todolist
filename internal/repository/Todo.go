package repository

import "time"

type Todo struct {
	UserId    int       `gorm:"not null;index"`
	Id        int       `gorm:"primaryKey;autoIncrement"`
	Titles    string    `gorm:"not null;size:255"`
	Done      bool      `gorm:"default:false;not null"`
	Content   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
