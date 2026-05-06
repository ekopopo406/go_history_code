package controllers

import (
	"errors"
	"go_chat_v1_test/internal/auth"
	"go_chat_v1_test/internal/exceptions"
	"go_chat_v1_test/internal/i18n"
	"go_chat_v1_test/internal/middlewares"
	"go_chat_v1_test/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authservice   services.AuthServices
	commonService services.CommonServices
}

func NewAuthController(authservice services.AuthServices, commonService services.CommonServices) *AuthController {
	return &AuthController{authservice: authservice, commonService: commonService}
}

type loginUserInput struct {
	UserEmail string `json:"userEmail" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
}

func (a *AuthController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.URL.Query().Get("lang")
		var inputs loginUserInput
		var respondMsg string
		ctx := c.Request.Context()
		// 解析 JSON 请求体
		if err := c.ShouldBindJSON(&inputs); err != nil {
			respondMsg = i18n.T(lang, "invalid_request")
			middlewares.RespondWithJSON(c, http.StatusBadRequest, respondMsg, nil)
			return
		}
		var responseData, err = a.authservice.Login(ctx, inputs.UserEmail, inputs.Password)
		if errors.Is(err, exceptions.ErrUserNotFoundInDB) {
			respondMsg = i18n.T(lang, exceptions.ErrUserNotFoundInDB.Error())
			middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, nil)
			return
		}

		middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, responseData)
	}

}

type loginUserRefreshInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (a *AuthController) Refresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.URL.Query().Get("lang")
		var inputs loginUserRefreshInput
		var respondMsg string

		ctx := c.Request.Context()

		if err := c.ShouldBindJSON(&inputs); err != nil {
			respondMsg = i18n.T(lang, "invalid_request")
			middlewares.RespondWithJSON(c, http.StatusBadRequest, respondMsg, nil)
			return
		}
		var responseData, err = a.authservice.Refresh(ctx, inputs.RefreshToken)
		if errors.Is(err, exceptions.ErrUserAlreadyLogout) {
			respondMsg = i18n.T(lang, "refresh_token_failed")
			middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, nil)
			return
		}
		respondMsg = i18n.T(lang, "success")
		middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, responseData)

	}

}

func (a *AuthController) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.URL.Query().Get("lang")

		var respondMsg string

		ctx := c.Request.Context()
		claims, ok := ctx.Value(middlewares.ClaimsKey).(*auth.Claims)

		if !ok {
			middlewares.RespondWithJSON(c, http.StatusUnauthorized, "未找到用户信息", nil)
			return
		}

		a.authservice.Logout(ctx, claims)
		respondMsg = i18n.T(lang, "success", "123123")
		middlewares.RespondWithJSON(c, http.StatusOK, respondMsg, respondMsg)
	}

}
