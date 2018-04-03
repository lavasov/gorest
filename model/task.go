package model

import "time"

type Task struct {
	ID          int64      `gorm:"primary_key" json:"id"`
	Title       string     `gorm:"size:255" json:"title"`
	Description string     `gorm:"size:255" json:"description"`
	Priority    int        `gorm:"type:int"`
	CreatedAt   time.Time  `gorm:"default:now()"`
	UpdatedAt   time.Time  `gorm:"default:now()"`
	CompletedAt *time.Time `gorm:"default:null"`
	IsDeleted   bool       `gorm:"default:false"`
	IsCompleted bool       `gorm:"default:false"`
	DeletedAt   *time.Time `gorm:"default:null"`
}
