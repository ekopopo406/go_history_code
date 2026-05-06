package repository

import (
	"context"
	apptypes "go_chat_v1_test/internal/types"

	"gorm.io/gorm"
)

type Entity interface {
	any
}

type Repository[T Entity] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]*T, error)
	Count(ctx context.Context) (int64, error)
	FindWithPagination(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*T, int64, error)
	DB() *gorm.DB
}

type Transactional interface {
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type GORMRepository[T Entity] struct {
	mainDb apptypes.MainDB
}

func NewGORMRepository[T Entity](mainDb apptypes.MainDB) Repository[T] {
	return &GORMRepository[T]{mainDb: mainDb}
}

func (r *GORMRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.mainDb.Db.WithContext(ctx).Create(entity).Error
}

func (r *GORMRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.mainDb.Db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GORMRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.mainDb.Db.WithContext(ctx).Save(entity).Error
}

func (r *GORMRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	return r.mainDb.Db.WithContext(ctx).Delete(&entity, id).Error
}

func (r *GORMRepository[T]) List(ctx context.Context) ([]*T, error) {
	var entities []*T
	err := r.mainDb.Db.WithContext(ctx).Find(&entities).Error
	return entities, err
}

func (r *GORMRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.mainDb.Db.WithContext(ctx).Model(&entity).Count(&count).Error
	return count, err
}

func (r *GORMRepository[T]) FindWithPagination(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*T, int64, error) {
	var entities []*T
	var total int64
	var entity T

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := r.mainDb.Db.WithContext(ctx).Model(&entity)

	// 简单过滤
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(pageSize).Find(&entities).Error

	return entities, total, err
}

// 事务支持
func (r *GORMRepository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.mainDb.Db.WithContext(ctx).Transaction(fn)
}

func (r *GORMRepository[T]) DB() *gorm.DB {
	return r.mainDb.Db
}
