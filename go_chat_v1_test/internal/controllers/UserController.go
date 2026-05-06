package controllers

import (
	"errors"
	"go_chat_v1_test/internal/common"
	"go_chat_v1_test/internal/exceptions"
	"go_chat_v1_test/internal/i18n"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/middlewares"
	"strings"

	"go_chat_v1_test/internal/services"
	"go_chat_v1_test/internal/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	userService  services.UserServices
	redisService services.RedisServices
	logger       *logger.Manager
}

func NewUserController(userService services.UserServices, redisService services.RedisServices, logger *logger.Manager) *UserController {
	return &UserController{userService: userService, redisService: redisService, logger: logger}
}

func (c *UserController) CreateUser(v *validator.Validator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Request.URL.Query().Get("lang")

		req := &common.CreateUserInput{}
		var respondMsg string
		// 解析并验证
		if err := v.ParseAndValidate(ctx.Request, req, lang); err != nil {
			// 格式化验证错误
			errors := v.FormatValidationErrors(err, lang)
			middlewares.RespondWithJSON(ctx, http.StatusBadRequest, errors, "")
			return
		}

		// 验证Redis随机码
		result, rediserr := c.redisService.Get(ctx.Request.Context(), "PHONE:"+req.PhoneNumber)
		if result != req.RandCode {
			respondMsg = i18n.GetErrorsMessage(lang, exceptions.ErrRandCodeExpire.Error())
			middlewares.RespondWithJSON(ctx, http.StatusBadRequest, respondMsg, "")
			return
		}
		if rediserr != nil {
			c.logger.Redis.Error("Error occurred while fetching random code from Redis. req.PhoneNumber:"+req.PhoneNumber, zap.Error(rediserr))
			middlewares.RespondWithJSON(ctx, http.StatusInternalServerError, "Internal server error", nil)
			return
		}
		user, err := c.userService.CreateUser(ctx, req)
		if errors.Is(err, exceptions.ErrDuplicPhone) {
			respondMsg = i18n.GetErrorsMessage(lang, exceptions.ErrDuplicPhone.Error())
			middlewares.RespondWithJSON(ctx, http.StatusOK, respondMsg, nil)
			return
		} else {
			respondMsg = i18n.T(lang, "success")
		}

		middlewares.RespondWithJSON(ctx, http.StatusOK, respondMsg, user)
	}
}

func (c *UserController) GetAllUserByPage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Request.URL.Query().Get("lang")
		req := &common.SearchUserInput{}
		var respondMsg string
		filters := make(map[string]any)
		var users, _, _ = c.userService.SearchUser(ctx, req, filters)
		respondMsg = i18n.T(lang, "success", "123123")
		middlewares.RespondWithJSON(ctx, http.StatusOK, respondMsg, users)
	}
}

func (c *UserController) GetAllUserByPage2(v *validator.Validator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Request.URL.Query().Get("lang")
		req := &common.SearchUserInput{}
		var respondMsg string
		// 2. 创建 JSON 解码器并解析请求体
		// decoder := json.NewDecoder(r.Body)
		// if err := decoder.Decode(&req); err != nil {
		// 	respondMsg = i18n.T(lang, "param_error", req)
		// 	http.Error(w, "Invalid JSON format: "+respondMsg, http.StatusBadRequest)
		// 	return
		// }
		//defer r.Body.Close() // 重要：记得关闭 Body

		// 解析并验证
		if err := v.ParseAndValidate(ctx.Request, &req, lang); err != nil {
			// 格式化验证错误
			errors := v.FormatValidationErrors(err, lang)
			middlewares.RespondWithJSON(ctx, http.StatusBadRequest, errors, "")
			return
		}
		filters := make(map[string]any)
		var users, _, _ = c.userService.SearchUser(ctx, req, filters)
		respondMsg = i18n.T(lang, "success", "123123")
		middlewares.RespondWithJSON(ctx, http.StatusOK, respondMsg, users)
	}
}
func (c *UserController) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		// if err != nil {
		// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
		// 	return
		// }

		// user, err := c.userService.GetUserByID(uint(id))
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusNotFound)
		// 	return
		// }

		// middlewares.RespondWithJSON(w, http.StatusOK, "SUCCESS", user)
	}
}

func UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// vars := mux.Vars(r)
		// id, err := strconv.Atoi(vars["id"])
		// if err != nil {
		// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
		// 	return
		// }

		// var user models.UserBasic
		// if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		// 	http.Error(w, "Invalid request body", http.StatusBadRequest)
		// 	return
		// }

		// updatedUser, err := models.UpdateUser(id, user)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// respondWithJSON(w, http.StatusOK, updatedUser)
	}
}

func (c *UserController) DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// vars := mux.Vars(r)
		// id, err := strconv.Atoi(vars["id"])
		// if err != nil {
		// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
		// 	return
		// }

		// if err := models.DeleteUser(id); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// respondWithJSON(w, http.StatusNoContent, nil)
	}
}

func (c *UserController) SubMessage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 订阅多个频道
		channels := []string{"chat.room.*", "notification.user.*", "system.broadcast"}

		err := c.redisService.StartSubScribe(ctx, channels,
			func(channel, payload string) {
				c.logger.Redis.Info("Received message",
					zap.String("channel", channel),
					zap.String("payload", payload))

				// 根据 channel 做不同业务处理
				switch {
				case strings.HasPrefix(channel, "chat.room."):
					c.logger.Redis.Info("Received chat.room. message",
						zap.String("channel", channel),
						zap.String("payload", payload))
				case strings.HasPrefix(channel, "notification."):
					c.logger.Redis.Info("Received notification. message",
						zap.String("payload", payload))
				}
			})

		if err != nil {
			c.logger.Redis.Fatal("Failed to start subscriber", zap.Error(err))
		}

	}
}

func (c *UserController) PubMessage() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 测试发布消息
		go func() {
			c.redisService.PublishMessage(ctx, "chat.room.1001", `{"user_id":123,"message":"Hello everyone!"}`)
		}()
	}
}

func HomePage() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		middlewares.RespondWithJSON(ctx, http.StatusOK, "SUCCESS", "yes! You Got it!")
	}
}
