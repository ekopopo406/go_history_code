package services

import (
	"context"
	"go_chat_v1_test/internal/common"
	"go_chat_v1_test/internal/dto"
)

type UserServices interface {

	// CreateUser 创建用户的业务逻辑
	CreateUser(ctx context.Context, input *common.CreateUserInput) (*dto.UserResponse, error)
	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, input *common.UpdateUserInput) (*dto.UserResponse, error)

	SearchUser(ctx context.Context, input *common.SearchUserInput, filters map[string]interface{}) ([]*dto.UserResponse, int64, error)

	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
}
