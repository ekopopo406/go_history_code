package chat

import (
	"encoding/json"
	"errors"
	"go_chat_v1_test/internal/utils"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) ReadPump(hub *Hub, onMessage func(*Message)) {
	defer func() {
		hub.Unregister(c, c.roomID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client %s read error: %v", c.id, err)
			}
			break
		}

		c.lastPing = time.Now()

		// 解析消息
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Client %s unmarshal error: %v", c.id, err)
			continue
		}

		// 填充发送者信息
		msg.SenderID = c.id
		msg.SenderUserID = c.userID
		msg.Timestamp = time.Now()
		if msg.ID == "" {
			msg.ID = utils.GenerateUUID()
		}

		// 处理心跳
		if msg.Type == "ping" {
			c.Send(&Message{
				ID:        utils.GenerateUUID(),
				Type:      "pong",
				Timestamp: time.Now(),
			})
			continue
		}

		if onMessage != nil {
			onMessage(&msg)
		}

		// 转发到房间
		if msg.RoomID != "" {
			hub.BroadcastToRoom(msg.RoomID, &msg)
		}
	}
}

// WritePump 写入循环（在每个连接的 goroutine 中运行）
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			if err := c.conn.WriteJSON(msg); err != nil {
				c.mu.Unlock()
				log.Printf("Client %s write error: %v", c.id, err)
				return
			}
			c.mu.Unlock()

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send 同步发送消息
func (c *Client) Send(msg *Message) error {
	select {
	case c.send <- msg:
		return nil
	case <-time.After(3 * time.Second):
		return errors.New("send timeout")
	}
}

// SendJSON 发送自定义 JSON
func (c *Client) SendJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(v)
}

// Close 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
}

// IsAlive 检查连接是否存活
func (c *Client) IsAlive(timeout time.Duration) bool {
	return time.Since(c.lastPing) < timeout
}
