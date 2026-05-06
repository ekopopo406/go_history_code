package chat

import (
	"fmt"
	"go_chat_v1_test/internal/utils"
	"time"
)

func (r *Room) Run() {
	for {
		select {
		case reg, ok := <-r.register:
			if !ok {
				return
			}
			r.mu.Lock()

			r.clients[reg.id] = reg
			r.lastActive = time.Now()
			r.mu.Unlock()
			sysMsg := &Message{
				ID:        utils.GenerateUUID(),
				Type:      "system",
				RoomID:    r.id,
				Content:   fmt.Sprintf("user %s joined", reg.userID),
				Timestamp: time.Now(),
			}
			r.broadcastToClients(sysMsg)
		case unreg, ok := <-r.unregister:
			if !ok {
				return
			}
			r.mu.Lock()
			if _, exists := r.clients[unreg.id]; exists {
				delete(r.clients, unreg.id)
				close(unreg.send)
				r.lastActive = time.Now()
			}
			r.mu.Unlock()

			sysMsg := &Message{
				ID:        utils.GenerateUUID(),
				Type:      "system",
				RoomID:    r.id,
				Content:   fmt.Sprintf("user %s left", unreg.userID),
				Timestamp: time.Now(),
			}
			r.broadcastToClients(sysMsg)
		case broadcast, ok := <-r.broadcast:
			if !ok {
				return
			}
			r.broadcastToClients(broadcast)
		}
	}
}

func (r *Room) GetMetadata() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m := make(map[string]interface{})
	for k, v := range r.metadata {
		m[k] = v
	}
	return m
}

// SetMetadata 设置Room元数据
func (r *Room) SetMetadata(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metadata[key] = value
}

// GetClientCount 获取客户端数量
func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// GetClients 获取所有客户端
func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.clients))
	for _, c := range r.clients {
		clients = append(clients, c)
	}
	return clients
}

func (r *Room) broadcastToClients(msg *Message) {
	r.mu.RLock()
	clients := make([]*Client, 0, len(r.clients))
	for _, c := range r.clients {
		clients = append(clients, c)
	}
	r.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- msg:
		default:
			r.mu.Lock()
			delete(r.clients, client.id)
			close(client.send)
			r.mu.Unlock()
			client.conn.Close()
		}
	}
}
