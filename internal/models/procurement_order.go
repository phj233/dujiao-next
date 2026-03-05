package models

import (
	"time"

	"gorm.io/gorm"
)

// ProcurementOrder 采购单表（A 向 B 发起的上游采购记录）
type ProcurementOrder struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	ConnectionID    uint           `gorm:"index;not null" json:"connection_id"`
	LocalOrderID    uint           `gorm:"index;not null" json:"local_order_id"`
	LocalOrderNo    string         `gorm:"type:varchar(64);index" json:"local_order_no"`
	UpstreamOrderID uint           `json:"upstream_order_id,omitempty"`
	UpstreamOrderNo string         `gorm:"type:varchar(64);index" json:"upstream_order_no,omitempty"`
	Status          string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	UpstreamAmount  Money          `gorm:"type:decimal(20,2);not null;default:0" json:"upstream_amount"`
	LocalSellAmount Money          `gorm:"type:decimal(20,2);not null;default:0" json:"local_sell_amount"`
	Currency        string         `gorm:"type:varchar(10);not null" json:"currency"`
	ErrorMessage    string         `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount      int            `gorm:"not null;default:0" json:"retry_count"`
	NextRetryAt     *time.Time     `gorm:"index" json:"next_retry_at,omitempty"`
	UpstreamPayload string         `gorm:"type:text" json:"upstream_payload,omitempty"`
	TraceID         string         `gorm:"type:varchar(64);index" json:"trace_id"`
	CreatedAt       time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Connection *SiteConnection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
	LocalOrder *Order          `gorm:"foreignKey:LocalOrderID" json:"local_order,omitempty"`
}

// TableName 指定表名
func (ProcurementOrder) TableName() string {
	return "procurement_orders"
}
