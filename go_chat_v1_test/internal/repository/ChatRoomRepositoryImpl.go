package repository

import (
	"errors"
	"go_chat_v1_test/internal/dto/models"

	"gorm.io/gorm"
)

// ChatRoomRepository 结构体
type ChatRoomRepositoryImpl struct {
	genericRepo Repository[models.ChatRoom]
}

// NewChatRoomRepository 构造函数
func NewChatRoomRepositoryImpl(genericRepo Repository[models.ChatRoom]) ChatRoomRepository {
	return &ChatRoomRepositoryImpl{genericRepo: genericRepo}
}

func (c *ChatRoomRepositoryImpl) CreateChatRoom(userID string, roomName string) (*models.ChatRoom, bool) {
	var existChatRoom *models.ChatRoom
	err := c.genericRepo.DB().Where("room_owner_user_id = ?", userID).First(&existChatRoom).Error

	if err != nil {
		return existChatRoom, true
	}
	tempChatRoom := &models.ChatRoom{
		RoomName:        roomName,
		RoomOwnerUserID: userID,
	}
	c.genericRepo.DB().Create(tempChatRoom)
	return tempChatRoom, true
}
func (c *ChatRoomRepositoryImpl) GetChatRoomByUserID(userID string) *models.ChatRoom {
	existChatRoom := &models.ChatRoom{}
	result := c.genericRepo.DB().Where("room_owner_user_id = ?", userID).First(existChatRoom)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return existChatRoom
}

func (c *ChatRoomRepositoryImpl) GetChatRoomByUserID2(userID string) *models.ChatRoom {
	existChatRoom := &models.ChatRoom{}
	result := c.genericRepo.DB().Where("room_owner_user_id = ?", userID).First(existChatRoom)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return existChatRoom
}
