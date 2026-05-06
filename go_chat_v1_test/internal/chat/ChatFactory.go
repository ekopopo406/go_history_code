package chat

import "github.com/gorilla/websocket"

type ClientFactory interface {
	NewClient(userID string, roomID string, hubID int, conn *websocket.Conn) *Client
}

type RoomFactory interface {
	NewRoom(roomID string, hubID int) *Room
}
