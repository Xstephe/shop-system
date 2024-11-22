package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32          `gorm:"primary_key;type:int" json:"id"`
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
	IsDeleted bool           `json:"-"`
}
