package chat

import (
	"fmt"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/services"
	"go_chat_v1_test/internal/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	shardedHub    *ShardedHub
	clientFactory ClientFactory
	roomFactory   RoomFactory
	redisServices services.RedisServices
	logger        *logger.Manager
}

func NewWebSocketHandler(shardedHub *ShardedHub, clientFactory ClientFactory, roomFactory RoomFactory, redisServices services.RedisServices, logger *logger.Manager) *WebSocketHandler {
	return &WebSocketHandler{shardedHub: shardedHub, clientFactory: clientFactory, roomFactory: roomFactory, redisServices: redisServices, logger: logger}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		log.Printf("检查来源: %s", origin)
		return true // 允许所有来源
	},
}

// ServeWs 处理 WebSocket 连接请求
func (h *WebSocketHandler) ServeWs(c *gin.Context) {
	h.logger.Ws.Info("收到 WebSocket 请求: %s", zap.String("remoteAddr", c.Request.RemoteAddr))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Ws.Error("升级失败: %v", zap.Error(err))
		return
	}
	userID := c.Query("user_id")
	clientID := utils.GenerateUUID()

	hubId := 1
	roomID := c.Query("room_id")
	fmt.Printf("clientID %s roomID %s", clientID, roomID)
	if userID == "" {
		h.logger.Ws.Warn("缺少 user_id")
		h.logger.Ws.Warn("缺少 client_id")
		conn.Close()
		return
	}
	client := h.clientFactory.NewClient(userID, clientID, hubId, conn)
	client.roomID = roomID
	//room := h.roomFactory.NewRoom(roomID, hubId)

	// 注册到 Hub
	hub := h.shardedHub.GetHub(hubId)
	if err := hub.Register(client, roomID); err != nil {
		log.Printf("Register client error: %v", err)
		conn.Close()
		return
	}

	// 只在 Register 成功后才启动 pump（Register 失败会 return）
	go client.WritePump()
	go client.ReadPump(hub, nil)

	h.logger.Ws.Info("WebSocket 连接建立完成: userID=%s", zap.String("userID", userID))
}
