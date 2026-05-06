package controllers

import (
	"go_chat_v1_test/internal/chat"
	"log"

	"github.com/gin-gonic/gin"
)

type WebSocketController struct {
	wsHandler *chat.WebSocketHandler
}

func NewWebSocketController(wsHandler *chat.WebSocketHandler) *WebSocketController {
	return &WebSocketController{
		wsHandler: wsHandler,
	}
}

func (h *WebSocketController) Getws() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🌐 收到 /ws 请求: %s", c.Request.RemoteAddr)
		h.wsHandler.ServeWs(c)
	}

}
