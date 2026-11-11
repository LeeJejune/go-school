package mysql

import (
	"context"

	"github.com/LeeJejune/go-school/go-practice/internal/catalog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ResourceEntity struct {
	ID    string `gorm:"column:id;primaryKey"`
	Title string `gorm:"column:title"`
	URL   string `gorm:"column:url"`
}

func (ResourceEntity) TableName() string { return "catalog_resources" }

type CatalogRepository struct{ db *gorm.DB }

var _ catalog.Repository = (*CatalogRepository)(nil)

func NewCatalogRepository(db *gorm.DB) *CatalogRepository { return &CatalogRepository{db: db} }

func (r *CatalogRepository) List(ctx context.Context) ([]catalog.Resource, error) {
	var rows []ResourceEntity
	if err := r.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}
	resources := make([]catalog.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, catalog.Resource{ID: row.ID, Title: row.Title, URL: row.URL})
	}
	return resources, nil
}
func (r *CatalogRepository) Batch(ctx context.Context, ids []string) ([]catalog.Resource, error) {
	var rows []ResourceEntity
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]catalog.Resource, len(rows))
	for _, row := range rows {
		byID[row.ID] = catalog.Resource{ID: row.ID, Title: row.Title, URL: row.URL}
	}
	resources := make([]catalog.Resource, 0, len(ids))
	for _, id := range ids {
		resource, ok := byID[id]
		if !ok {
			return nil, catalog.ErrNotFound
		}
		resources = append(resources, resource)
	}
	return resources, nil
}
func (r *CatalogRepository) Upsert(ctx context.Context, resource catalog.Resource) (catalog.Resource, error) {
	row := ResourceEntity{ID: resource.ID, Title: resource.Title, URL: resource.URL}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"title", "url"})}).Create(&row).Error
	return resource, err
}
