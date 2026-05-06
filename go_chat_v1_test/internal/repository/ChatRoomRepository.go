package repository

import "go_chat_v1_test/internal/dto/models"

type ChatRoomRepository interface {
	CreateChatRoom(userID string, roomName string) (*models.ChatRoom, bool)
	GetChatRoomByUserID(userID string) *models.ChatRoom
}
