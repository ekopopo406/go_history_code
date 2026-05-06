package services

import "go_chat_v1_test/internal/dto/models"

type ChatServices interface {
	createChatRoom(userID string, roomName string) bool
	getChatRoom(userID string) *models.ChatRoom
}
