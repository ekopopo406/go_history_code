package chat

import (
	"go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/utils"
	"time"

	"github.com/gorilla/websocket"
)

type DefaultClientFactory struct {
	logger *logger.Manager
	// redis *redis.Client
	config *config.WebSocketConfig
	// 其他你想注入给每个 Client 的东西
}

type DefaultRoomFactory struct {
	logger *logger.Manager
	// redis *redis.Client
	config *config.WebSocketConfig
	// 其他你想注入给每个 Room 的东西
}

func NewDefaultClientFactory(logger *logger.Manager, config *config.WebSocketConfig) ClientFactory {
	return &DefaultClientFactory{
		logger: logger,
		config: config,
	}
}

func NewDefaultRoomFactory(logger *logger.Manager, config *config.WebSocketConfig) RoomFactory {
	return &DefaultRoomFactory{
		logger: logger,
		config: config,
	}
}

func (f *DefaultClientFactory) NewClient(userID string, roomID string, hubID int, conn *websocket.Conn) *Client {
	return &Client{
		id:        utils.GenerateUUID(),
		userID:    userID,
		roomID:    roomID,
		hubID:     hubID,
		conn:      conn,
		send:      make(chan *Message, 50),
		createdAt: time.Now(),
		lastPing:  time.Now(),
		logger:    f.logger,
		config:    f.config,
	}
}

func (f *DefaultRoomFactory) NewRoom(roomID string, hubID int) *Room {
	return &Room{
		id:         roomID,
		hubID:      hubID,
		clients:    make(map[string]*Client, 100),
		broadcast:  make(chan *Message, 100),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		createdAt:  time.Now(),
		lastActive: time.Now(),
		metadata:   make(map[string]interface{}),
	}
}
