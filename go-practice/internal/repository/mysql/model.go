// Package mysql contains persistence adapters and GORM entities.
package mysql

import "time"

// These structs map DB rows. They are intentionally separate from protobuf DTOs
// and the business models in internal/reading.
type ListEntity struct {
	ID          string       `gorm:"column:id;primaryKey"`
	RequestKey  string       `gorm:"column:request_key"`
	Fingerprint string       `gorm:"column:fingerprint"`
	Title       string       `gorm:"column:title"`
	CreatedAt   time.Time    `gorm:"column:created_at;autoCreateTime:false;default:CURRENT_TIMESTAMP(6)"`
	Items       []ItemEntity `gorm:"foreignKey:ListID;references:ID"`
}

func (ListEntity) TableName() string { return "reading_lists" }

type ItemEntity struct {
	ListID     string `gorm:"column:list_id;primaryKey"`
	ResourceID string `gorm:"column:resource_id;primaryKey"`
	Title      string `gorm:"column:title"`
	URL        string `gorm:"column:url"`
	Position   int    `gorm:"column:position"`
	IsRead     bool   `gorm:"column:is_read"`
}

func (ItemEntity) TableName() string { return "reading_items" }
