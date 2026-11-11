package mysql

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/LeeJejune/go-school/go-practice/internal/reading"
	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type ReadingRepository struct{ db *gorm.DB }

var _ reading.Repository = (*ReadingRepository)(nil)

func NewReadingRepository(db *gorm.DB) *ReadingRepository { return &ReadingRepository{db: db} }

func (r *ReadingRepository) Create(ctx context.Context, params reading.CreateParams) (reading.List, error) {
	if len(params.Resources) == 0 {
		return reading.List{}, errors.New("a reading list needs at least one resource")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return reading.List{}, err
	}
	entity := ListEntity{ID: hex.EncodeToString(id), RequestKey: params.RequestKey, Fingerprint: params.Fingerprint, Title: params.Title}
	duplicateRequest := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items").Create(&entity).Error; err != nil {
			var duplicate *mysqldriver.MySQLError
			duplicateRequest = errors.As(err, &duplicate) && duplicate.Number == 1062
			return fmt.Errorf("create list: %w", err)
		}
		items := make([]ItemEntity, 0, len(params.Resources))
		for position, resource := range params.Resources {
			items = append(items, ItemEntity{ListID: entity.ID, ResourceID: resource.ID, Title: resource.Title, URL: resource.URL, Position: position})
		}
		if err := tx.Create(&items).Error; err != nil {
			return fmt.Errorf("create items: %w", err)
		}
		return nil // GORM commits on nil, rolls back on an error or panic.
	})
	// Replay only a duplicate list key, never an item uniqueness failure. The
	// failed transaction has ended before another connection is requested.
	if duplicateRequest {
		return r.replay(ctx, params)
	}
	if err != nil {
		return reading.List{}, err
	}
	return r.Get(ctx, entity.ID)
}

func (r *ReadingRepository) replay(ctx context.Context, params reading.CreateParams) (reading.List, error) {
	var entity ListEntity
	if err := r.db.WithContext(ctx).Where("request_key = ?", params.RequestKey).First(&entity).Error; err != nil {
		return reading.List{}, err
	}
	if entity.Fingerprint != params.Fingerprint {
		return reading.List{}, reading.ErrConflict
	}
	return r.Get(ctx, entity.ID)
}

func (r *ReadingRepository) Get(ctx context.Context, id string) (reading.List, error) {
	var entity ListEntity
	err := r.db.WithContext(ctx).Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return reading.List{}, reading.ErrNotFound
	}
	if err != nil {
		return reading.List{}, err
	}
	list := reading.List{ID: entity.ID, Title: entity.Title, CreatedAt: entity.CreatedAt}
	for _, item := range entity.Items {
		list.Items = append(list.Items, reading.Item{Resource: reading.Resource{ID: item.ResourceID, Title: item.Title, URL: item.URL}, IsRead: item.IsRead})
	}
	return list, nil
}

func (r *ReadingRepository) SetRead(ctx context.Context, listID, resourceID string, isRead bool) error {
	query := r.db.WithContext(ctx).Model(&ItemEntity{}).Where("list_id = ? AND resource_id = ?", listID, resourceID)
	result := query.Update("is_read", isRead) // Update includes false, unlike Updates(struct).
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	// MySQL counts changed rows; repeating the same value must still succeed.
	var count int64
	if err := r.db.WithContext(ctx).Model(&ItemEntity{}).Where("list_id = ? AND resource_id = ?", listID, resourceID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return reading.ErrNotFound
	}
	return nil
}
