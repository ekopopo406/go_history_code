package repository

import (
	"context"
	"errors"
	"go_chat_v1_test/internal/dto"
	"go_chat_v1_test/internal/dto/models"

	"go_chat_v1_test/internal/exceptions"

	"gorm.io/gorm"
)

// UserRepository 结构体
type UserRepositoryImpl struct {
	genericRepo Repository[models.UserBasic]
}

// NewUserRepository 构造函数
func NewUserRepositoryImpl(genericRepo Repository[models.UserBasic]) UserRepository {
	return &UserRepositoryImpl{genericRepo: genericRepo}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.UserBasic) error {
	return r.genericRepo.Create(ctx, user)
}

// FindByID 根据ID查找
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.UserBasic, error) {
	var user models.UserBasic
	err := r.genericRepo.DB().First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 返回 nil 表示未找到
	}
	return &user, err
}

// FindByEmail 根据邮箱查找
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.UserBasic, error) {
	var user models.UserBasic
	err := r.genericRepo.DB().Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exceptions.ErrUserNotFoundInDB // 返回自定义错误
		}
		return nil, err
	}
	return &user, err
}

// FindByPhone 根据电话查找
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, email string) (int64, error) {
	var count int64
	err := r.genericRepo.DB().Model(&dto.UserResponse{}).
		Where("phone = ?", email).
		Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, exceptions.ErrUserNotFoundInDB // 返回自定义错误
		}
		return 0, err
	}
	if count > 0 {
		return count, exceptions.ErrDuplicPhone // 返回自定义错误
	}
	return count, err
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.UserBasic) error {
	return r.genericRepo.Update(ctx, user)
}

// Update 更新用户密码
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, user *models.UserBasic) error {
	return r.genericRepo.Update(ctx, user)
}

// Delete 删除用户（软删除）
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.genericRepo.Delete(ctx, id)
}

// List 用户列表
func (r *UserRepositoryImpl) FindWithPagination(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*models.UserBasic, int64, error) {

	users, total, err := r.genericRepo.FindWithPagination(ctx, page, pageSize, filters)

	return users, total, err
}
