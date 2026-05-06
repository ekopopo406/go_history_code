package impl

import (
	"context"
	"errors"
	"go_chat_v1_test/internal/common"
	"go_chat_v1_test/internal/dto"
	"go_chat_v1_test/internal/dto/models"
	"go_chat_v1_test/internal/repository"
	"go_chat_v1_test/internal/services"

	"golang.org/x/crypto/bcrypt"
)

// UserService 业务逻辑层
type UserServiceImpl struct {
	userRepo     repository.UserRepository
	kafkaService services.KafkaServices
}

// NewUserService 构造函数
func NewUserServiceImpl(userRepo repository.UserRepository, kafkaService services.KafkaServices) services.UserServices {
	return &UserServiceImpl{userRepo: userRepo, kafkaService: kafkaService}
}

// CreateUser 创建用户的业务逻辑
func (s *UserServiceImpl) CreateUser(ctx context.Context, input *common.CreateUserInput) (*dto.UserResponse, error) {
	// 1. 验证输入
	if input.Name == "" {
		return nil, errors.New("用户名不能为空")
	}

	// 2. 检查电话是否已存在
	existingUser, err := s.userRepo.FindByPhone(ctx, input.PhoneNumber)
	if existingUser != 0 {
		return nil, err
	}

	// 3. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. 创建用户对象
	user := &models.UserBasic{
		Name:     input.Name,
		Phone:    input.PhoneNumber,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	// 5. 保存到数据库
	err = s.userRepo.Create(ctx, user)
	return dto.ToUserResponse(user), err
}

// UpdateUserInput 更新用户输入
type UpdateUserInput struct {
	ID   uint
	Name string
	Age  int
}

// UpdateUser 更新用户信息
func (s *UserServiceImpl) UpdateUser(ctx context.Context, input *common.UpdateUserInput) (*dto.UserResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByID(ctx, input.ID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}

	// 2. 更新字段
	if input.Name != "" {
		user.Name = input.Name
	}

	// 3. 保存
	err = s.userRepo.Update(ctx, user)
	return dto.ToUserResponse(user), err
}

func (s *UserServiceImpl) SearchUser(ctx context.Context, input *common.SearchUserInput, filters map[string]interface{}) ([]*dto.UserResponse, int64, error) {

	// 1. 查找用户
	users, total, err := s.userRepo.FindWithPagination(ctx, input.Page, input.PageSize, filters)
	if err != nil || users == nil {
		return nil, 0, errors.New("用户不存在")
	}

	return dto.ToUserResponseList(users), total, err
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {

	// 1. 查找用户
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}

	return dto.ToUserResponse(user), err
}
