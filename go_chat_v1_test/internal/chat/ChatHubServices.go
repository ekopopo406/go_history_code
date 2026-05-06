package chat

import (
	"errors"
	"fmt"
	"log"
	"time"
)

func (h *Hub) Run() {
	fmt.Println("HUB start Running")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case reg := <-h.register:
			fmt.Println("HUB register")
			h.mu.RLock()
			room, exists := h.rooms[reg.RoomID]
			h.mu.RUnlock()

			if !exists {
				room, _ = h.CreateRoom(reg.RoomID)
				fmt.Printf("HUB register and create new room with ROOMID %s", reg.RoomID)
				h.mu.Lock()
				h.rooms[reg.RoomID] = room
				h.mu.Unlock()
			}
			room.register <- reg.Client
		case unreg := <-h.unregister:
			fmt.Println("HUB unregister")
			//delete(h.rooms, unreg.RoomID)
			h.mu.RLock()
			room, exists := h.rooms[unreg.RoomID]
			h.mu.RUnlock()

			if exists {
				room.unregister <- unreg.Client
			}
		case brodcast := <-h.broadcast:

			h.mu.RLock()
			room, exists := h.rooms[brodcast.RoomID]
			h.mu.RUnlock()
			if exists {
				room.broadcast <- brodcast
			}
		case <-ticker.C:
			h.cleanupInactiveRooms()

		case <-h.stop:
			h.cleanup()
			return

		}
	}

}

// SetRoomFactory 设置房间工厂
func (h *Hub) SetRoomFactory(factory RoomFactory) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.roomFactory = factory
}

// SetClientFactory 设置客户端工厂
func (h *Hub) SetClientFactory(factory ClientFactory) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clientFactory = factory
}

// Register 注册客户端
func (h *Hub) Register(client *Client, roomID string) error {
	if client == nil {
		return errors.New("client is nil")
	}

	client.roomID = roomID
	client.hubID = h.id

	select {
	case h.register <- &RegisterMsg{Client: client, RoomID: roomID}:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("register timeout")
	}
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client, roomID string) {
	if client == nil {
		return
	}
	select {
	case h.unregister <- &UnregisterMsg{Client: client, RoomID: roomID}:
	case <-time.After(2 * time.Second):
		log.Printf("Hub %s: unregister timeout for client %s", h.id, client.id)
	}
}

// Broadcast 向本分片所有房间广播
func (h *Hub) Broadcast(msg *Message, opts ...BroadcastOption) {
	select {
	case h.broadcast <- msg:
	case <-time.After(3 * time.Second):
		log.Printf("Hub %s: broadcast timeout", h.id)
	}
}

// BroadcastToRoom 向指定房间广播
func (h *Hub) BroadcastToRoom(roomID string, msg *Message, opts ...BroadcastOption) error {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	var opt BroadcastOption
	if len(opts) > 0 {
		opt = opts[0]
	}

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.clients))
	for _, c := range room.clients {
		// 应用过滤
		if opt.ExcludeClientID != "" && c.id == opt.ExcludeClientID {
			continue
		}
		if len(opt.ExcludeUserIDs) > 0 {
			excluded := false
			for _, uid := range opt.ExcludeUserIDs {
				if c.userID == uid {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		if opt.Filter != nil && !opt.Filter(c) {
			continue
		}
		clients = append(clients, c)
	}
	room.mu.RUnlock()

	// 异步发送
	for _, c := range clients {
		select {
		case c.send <- msg:
		default:
			log.Printf("Hub %d: client %s send buffer full", h.id, c.id)
		}
	}

	return nil
}

func (h *Hub) CreateRoom(roomID string) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[roomID]; exists {
		return nil, fmt.Errorf("room %s already exists", roomID)
	}

	room := h.roomFactory.NewRoom(roomID, h.id)
	h.rooms[roomID] = room

	h.wg.Add(1)
	go room.Run()
	h.wg.Done()

	return room, nil
}

func (h *Hub) DeleteRoom(roomID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return fmt.Errorf("room %s not found", roomID)
	}

	// 踢出所有客户端
	room.mu.Lock()
	for _, client := range room.clients {
		close(client.send)
		client.conn.Close()
	}
	room.clients = make(map[string]*Client)
	room.mu.Unlock()

	close(room.broadcast)
	close(room.register)
	close(room.unregister)
	delete(h.rooms, roomID)
	return nil
}

// GetRoom 获取房间
func (h *Hub) GetRoom(roomID string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}
	return room, nil
}

// ListRooms 列出所有房间
func (h *Hub) ListRooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]*Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (h *Hub) cleanupInactiveRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for roomID, room := range h.rooms {
		room.mu.RLock()
		clientCount := len(room.clients)
		lastActive := room.lastActive
		room.mu.RUnlock()

		// 空房间且超过 10 分钟未活跃则删除
		if clientCount == 0 && now.Sub(lastActive) > 10*time.Minute {
			close(room.broadcast)
			close(room.register)
			close(room.unregister)
			delete(h.rooms, roomID)
			log.Printf("Hub %d: cleaned up inactive room %s", h.id, roomID)
		}
	}
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, room := range h.rooms {
		close(room.broadcast)
		close(room.register)
		close(room.unregister)
	}
	h.rooms = make(map[string]*Room)
}
func (h *Hub) GetStats() HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := HubStats{
		HubID:       h.id,
		RoomCount:   len(h.rooms),
		ClientCount: 0,
		RoomDetails: make([]RoomStats, 0, len(h.rooms)),
	}

	for _, room := range h.rooms {
		room.mu.RLock()
		clientCount := len(room.clients)
		roomStats := RoomStats{
			RoomID:      room.id,
			ClientCount: clientCount,
			CreatedAt:   room.createdAt,
			LastActive:  room.lastActive,
		}
		room.mu.RUnlock()

		stats.ClientCount += clientCount
		stats.RoomDetails = append(stats.RoomDetails, roomStats)
	}

	return stats
}

func (h *Hub) Stop() {
	close(h.stop)
	h.wg.Wait()
}
