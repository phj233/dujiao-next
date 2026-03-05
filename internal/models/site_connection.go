package models

import (
	"time"

	"gorm.io/gorm"
)

// SiteConnection 对接连接表
type SiteConnection struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Name           string         `gorm:"type:varchar(100);not null" json:"name"`
	BaseURL        string         `gorm:"type:varchar(500);not null" json:"base_url"`
	ApiKey         string         `gorm:"type:varchar(64);not null" json:"api_key"`
	ApiSecret      string         `gorm:"type:varchar(256);not null" json:"-"` // AES-256 加密存储
	Protocol       string         `gorm:"type:varchar(20);not null;default:'dujiao-next'" json:"protocol"`
	CallbackURL    string         `gorm:"type:varchar(500)" json:"callback_url"`
	Status         string         `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	LastPingAt     *time.Time     `json:"last_ping_at,omitempty"`
	LastPingOK     bool           `gorm:"not null;default:false" json:"last_ping_ok"`
	RetryMax       int            `gorm:"not null;default:5" json:"retry_max"`
	RetryIntervals string         `gorm:"type:varchar(200);not null;default:'[30,60,300]'" json:"retry_intervals"`
	CreatedAt      time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (SiteConnection) TableName() string {
	return "site_connections"
}
