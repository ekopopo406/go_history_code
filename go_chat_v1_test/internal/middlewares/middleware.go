package middlewares

import (
	"context"
	"go_chat_v1_test/internal/auth"
	"go_chat_v1_test/internal/services"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				log.Printf(" recovered panic: %v\n%s", err, stack)
				// 如果用 zap：
				// zap.L().Error("recovered from panic",
				//     zap.Any("panic", rec),
				//     zap.String("stack", stack),
				// )

				// 可选：记录 request 信息，帮助定位
				log.Printf("Request: %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
		}()
		c.Next()
	}
}

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取语言，默认中文
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "zh-CN"
		}

		// 将语言存入请求上下文（通过URL参数）
		q := c.Request.URL.Query()
		q.Set("lang", lang)
		c.Request.URL.RawQuery = q.Encode()

		c.Next()
	}
}

type contextKey string

const UserIDKey contextKey = "userID"
const ClaimsKey contextKey = "claims"

// Middleware 创建一个 HTTP 中间件，验证 JWT 并检查 Redis 黑名单
func JwtMiddleware(accessSecret string, authService services.AuthServices) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 2. 解析 JWT
		claims, err := auth.ParseToken(tokenString, accessSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 3. 检查 Redis 黑名单
		blacklisted, err := authService.IsBlacklisted(c, claims.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			c.Abort()
			return
		}
		if blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文，传递给后续 Handler
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func SafeLogError(err error) {
	if err != nil {
		log.Printf("错误: %s, 时间: %v", err.Error(), time.Now())
	} else {
		log.Printf("错误: nil, 时间: %v", time.Now())
	}
}

// 辅助函数
//func RespondWithJSON(c *gin.Context, code int, message string, payload interface{}) {
// c.Writer.Header().Set("Content-Type", "application/json")
// c.Writer.WriteHeader(code)

// var res = &commonResponse{message, code, payload}
// json.NewEncoder(c.Writer).Encode(res)
//
//	c.JSON(code, gin.H{
//		"code":    code,
//		"message": message,
//		"data":    payload,
//	})
//
// }

type commonResponse struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

func RespondWithJSON(c *gin.Context, code int, message string, payload interface{}) {
	c.JSON(code, commonResponse{
		Message: message,
		Code:    code,
		Data:    payload,
	})
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
