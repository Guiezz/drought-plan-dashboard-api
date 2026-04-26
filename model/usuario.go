package model

import "time"

type Usuario struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nome      string    `gorm:"not null" json:"nome"`
	Email     string    `gorm:"unique;not null" json:"email"`
	SenhaHash string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:'cogerh'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Usuario) TableName() string {
	return "usuarios"
}
