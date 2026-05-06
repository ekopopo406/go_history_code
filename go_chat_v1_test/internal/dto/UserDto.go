package dto

import (
	"go_chat_v1_test/internal/dto/models"
	"time"
)

// UserResponse 用户响应结构体（不含敏感信息）
type UserResponse struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	ClientIP      string     `json:"client_ip,omitempty"`
	IdentityID    string     `json:"identity_id,omitempty"`
	ClientPort    string     `json:"client_port,omitempty"`
	LoginTime     *time.Time `json:"login_time,omitempty"`
	HeartBeatTime *time.Time `json:"heart_beat_time,omitempty"`
	LogoutTime    *time.Time `json:"logout_time,omitempty"`
	IsLogout      bool       `json:"is_logout"`
	DeviceInfo    string     `json:"device_info,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ToUserResponse 将 UserBasic 转换为 UserResponse
func ToUserResponse(user *models.UserBasic) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Phone:         user.Phone,
		Email:         user.Email,
		ClientIP:      user.ClientIP,
		IdentityID:    user.IdentityID,
		ClientPort:    user.ClientPort,
		LoginTime:     user.LoginTime,
		HeartBeatTime: user.HeartBeatTime,
		LogoutTime:    user.LogoutTime,
		IsLogout:      user.IsLogout,
		DeviceInfo:    user.DeviceInfo,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

// ToUserResponseList 批量转换
func ToUserResponseList(users []*models.UserBasic) []*UserResponse {
	responses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, ToUserResponse(user))
	}
	return responses
}
