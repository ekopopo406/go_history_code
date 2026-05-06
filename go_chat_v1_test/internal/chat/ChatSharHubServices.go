package chat

import (
	"fmt"
	"sync"
)

func NewShardedHub(maxHubs int) *ShardedHub {
	if maxHubs <= 0 {
		maxHubs = 8
	}

	sh := &ShardedHub{
		hubs:    make([]*Hub, maxHubs),
		hubSize: maxHubs,
	}

	for i := 0; i < maxHubs; i++ {

		h := &Hub{
			id:         i + 1,
			rooms:      make(map[string]*Room, 100),
			broadcast:  make(chan *Message, 100),
			register:   make(chan *RegisterMsg, 100),
			unregister: make(chan *UnregisterMsg, 100),
			stop:       make(chan struct{}),
		}

		sh.hubs[i] = h
		fmt.Printf("HUB START %d", i)
		sh.wg.Add(1)
		go h.Run()
		sh.wg.Done()
	}

	return sh
}

// GetAllHubs 获取所有分片
func (sh *ShardedHub) GetAllHubs() []*Hub {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	hubs := make([]*Hub, len(sh.hubs))
	copy(hubs, sh.hubs)
	return hubs
}

// GetHub 根据 roomID 获取对应分片 Hub
func (sh *ShardedHub) GetHub(hubID int) *Hub {
	return sh.hubs[hubID]
}

// Register 注册客户端到对应分片
func (sh *ShardedHub) Register(hubID int, client *Client, roomID string) error {
	hub := sh.GetHub(hubID)
	return hub.Register(client, roomID)
}

// Unregister 注销客户端
func (sh *ShardedHub) Unregister(hubID int, client *Client, roomID string) {
	hub := sh.GetHub(hubID)
	hub.Unregister(client, roomID)
}

// BroadcastToRoom 向指定房间广播
func (sh *ShardedHub) BroadcastToRoom(hubID int, roomID string, msg *Message, opts ...BroadcastOption) error {
	hub := sh.GetHub(hubID)
	return hub.BroadcastToRoom(roomID, msg, opts...)
}

// BroadcastToAll 向所有分片广播
func (sh *ShardedHub) BroadcastToAll(msg *Message, opts ...BroadcastOption) {
	var wg sync.WaitGroup
	for _, hub := range sh.hubs {
		wg.Add(1)
		go func(h *Hub) {
			defer wg.Done()
			h.Broadcast(msg, opts...)
		}(hub)
	}
	wg.Wait()
}

// CreateRoom 创建房间
func (sh *ShardedHub) CreateRoom(hubID int, roomID string) (*Room, error) {
	hub := sh.GetHub(hubID)
	return hub.CreateRoom(roomID)
}

// DeleteRoom 删除房间
func (sh *ShardedHub) DeleteRoom(hubID int, roomID string) error {
	hub := sh.GetHub(hubID)
	return hub.DeleteRoom(roomID)
}

// GetRoom 获取房间
func (sh *ShardedHub) GetRoom(hubID int, roomID string) (*Room, error) {
	hub := sh.GetHub(hubID)
	return hub.GetRoom(roomID)
}

// GetStats 获取所有分片统计
func (sh *ShardedHub) GetStats() []HubStats {
	stats := make([]HubStats, 0, sh.hubSize)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, hub := range sh.hubs {
		wg.Add(1)
		go func(h *Hub) {
			defer wg.Done()
			s := h.GetStats()
			mu.Lock()
			stats = append(stats, s)
			mu.Unlock()
		}(hub)
	}
	wg.Wait()
	return stats
}

// Stop 停止所有分片
func (sh *ShardedHub) Stop() {
	var wg sync.WaitGroup
	for _, hub := range sh.hubs {
		wg.Add(1)
		go func(h *Hub) {
			defer wg.Done()
			h.Stop()
		}(hub)
	}
	wg.Wait()
}
