package controllers

import (
	"go_chat_v1_test/internal/i18n"
	"go_chat_v1_test/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.URL.Query().Get("lang")
		var respondMsg string

		respondMsg = i18n.T(lang, "success")

		middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, "Pong")
	}

}
