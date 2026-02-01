package model

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:50"`
	Password  string `gorm:"size:255"` // 已哈希
	CreatedAt time.Time
}

type Submission struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Language  string `gorm:"size:20"`
	Code      string `gorm:"type:text"`
	Input     string `gorm:"type:text"` // 标准输入数据
	Status    string `gorm:"size:20"`   // Pending, Running, Success, Failed, Timeout
	Output    string `gorm:"type:text"`
	CreatedAt time.Time
}

// Language 表示支持的编程语言
type Language struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Value        string `gorm:"uniqueIndex;size:20" json:"value"` // 如: cpp, python, java
	Label        string `gorm:"size:50" json:"label"`             // 如: C++ (g++17)
	Icon         string `gorm:"size:100" json:"icon"`             // 图标文件名
	DisplayOrder int    `gorm:"default:0" json:"displayOrder"`    // 排序顺序
	Enabled      bool   `gorm:"default:true" json:"enabled"`      // 是否启用
}
