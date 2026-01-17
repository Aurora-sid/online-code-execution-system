package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;size:50"`
	Password  string    `gorm:"size:255"` // Hashed
	CreatedAt time.Time
}

type Submission struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Language  string    `gorm:"size:20"`
	Code      string    `gorm:"type:text"`
	Status    string    `gorm:"size:20"` // Pending, Running, Success, Failed, Timeout
	Output    string    `gorm:"type:text"`
	CreatedAt time.Time
}
