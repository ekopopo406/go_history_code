package chat

import (
	"go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ============= 1. 主分片管理器 (Main ShardedHub) =============
type ShardedHub struct {
	hubs    []*Hub           // 分片 Hub 列表 (Hub 1,2,3,4)
	hashFn  func(string) int // 分片哈希函数
	mu      sync.RWMutex     // 保护配置变更
	hubSize int              // 分片数量
	wg      sync.WaitGroup
}

// ============= 2. 单个分片 Hub (Hub 1/2/3/4) =============
type Hub struct {
	id            int                 // 分片 ID (1,2,3,4)
	rooms         map[string]*Room    // 房间映射: roomID -> Room
	mu            sync.RWMutex        // 保护 rooms map
	broadcast     chan *Message       // 广播通道（向本分片所有房间广播）
	register      chan *RegisterMsg   // 注册通道（客户端加入房间）
	unregister    chan *UnregisterMsg // 注销通道（客户端离开房间）
	stop          chan struct{}       // 停止信号
	wg            sync.WaitGroup      // 等待所有 goroutine 完成
	roomFactory   RoomFactory
	clientFactory ClientFactory
}

// ============= 3. 房间 (Room 1/2/3/4) =============
type Room struct {
	id         string                 // 房间 ID
	hubID      int                    // 所属分片 ID
	clients    map[string]*Client     // 客户端映射: clientID -> Client (Client 1-9999)
	broadcast  chan *Message          // 房间广播通道
	register   chan *Client           // 房间内注册
	unregister chan *Client           // 房间内注销
	mu         sync.RWMutex           // 保护 clients map
	createdAt  time.Time              // 创建时间
	lastActive time.Time              // 最后活跃时间
	metadata   map[string]interface{} // 房间元数据
}

// ============= 4. 客户端连接 (Client 1-9999) =============
type Client struct {
	id        string                 // 客户端唯一 ID
	userID    string                 // 用户 ID
	roomID    string                 // 当前所在房间 ID
	hubID     int                    // 所属分片 ID
	conn      *websocket.Conn        // WebSocket 连接
	send      chan *Message          // 发送消息通道（带缓冲）
	lastPing  time.Time              // 最后心跳时间
	mu        sync.Mutex             // 保护连接写入
	metadata  map[string]interface{} // 客户端元数据
	createdAt time.Time              // 连接建立时间
	logger    *logger.Manager
	config    *config.WebSocketConfig
}

// ============= 5. 消息结构 =============
type Message struct {
	ID           string                 `json:"id"`                 // 消息 ID
	Type         string                 `json:"type"`               // 消息类型: text, image, system, etc.
	RoomID       string                 `json:"room_id"`            // 目标房间 ID
	SenderID     string                 `json:"sender_id"`          // 发送者客户端 ID
	SenderUserID string                 `json:"sender_user_id"`     // 发送者用户 ID
	SenderName   string                 `json:"sender_name"`        // 发送者昵称
	Content      string                 `json:"content"`            // 消息内容
	Data         []byte                 `json:"-"`                  // 二进制数据
	Timestamp    time.Time              `json:"timestamp"`          // 消息时间戳
	Metadata     map[string]interface{} `json:"metadata,omitempty"` // 扩展字段
}

// ============= 6. 辅助结构体 =============

// 注册消息（用于 Hub 的 register 通道）
type RegisterMsg struct {
	Client *Client
	RoomID string
}

// 注销消息（用于 Hub 的 unregister 通道）
type UnregisterMsg struct {
	Client *Client
	RoomID string
}

// 广播选项
type BroadcastOption struct {
	ExcludeClientID string             // 排除的客户端 ID
	ExcludeUserIDs  []string           // 排除的用户 ID 列表
	Filter          func(*Client) bool // 自定义过滤函数
}

// 房间操作
type RoomOperation struct {
	Op       string // "create", "delete", "get", "list"
	RoomID   string
	Response chan interface{} // 用于返回结果
}

// 统计信息
type HubStats struct {
	HubID       int         `json:"hub_id"`
	RoomCount   int         `json:"room_count"`
	ClientCount int         `json:"client_count"`
	RoomDetails []RoomStats `json:"room_details,omitempty"`
}

type RoomStats struct {
	RoomID      string    `json:"room_id"`
	ClientCount int       `json:"client_count"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
}
