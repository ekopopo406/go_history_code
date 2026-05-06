package repository

import (
	"context"
	"go_chat_v1_test/internal/dto/models"
)

type UserRepository interface {

	// Create 创建用户
	Create(ctx context.Context, user *models.UserBasic) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uint) (*models.UserBasic, error)

	// FindByEmail 根据邮箱查找
	FindByEmail(ctx context.Context, email string) (*models.UserBasic, error)

	// FindByPhone 根据电话查找
	FindByPhone(ctx context.Context, email string) (int64, error)

	// Update 更新用户
	Update(ctx context.Context, user *models.UserBasic) error

	// Delete 删除用户（软删除）
	Delete(ctx context.Context, id uint) error

	// List 用户列表
	FindWithPagination(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*models.UserBasic, int64, error)
}
