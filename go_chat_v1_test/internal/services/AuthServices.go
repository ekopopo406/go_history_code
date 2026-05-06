package services

import (
	"context"
	"go_chat_v1_test/internal/auth"
	"time"
)

type AuthServices interface {

	// Login 登录逻辑
	Login(ctx context.Context, userEmail string, password string) (*auth.TokenPair, error)

	// Logout 登出逻辑
	Logout(ctx context.Context, claims *auth.Claims) error

	// Refresh 刷新令牌
	Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	// BlacklistToken 将 token 的 jti 加入黑名单，有效期与 token 剩余时间一致
	BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error

	// IsRefreshTokenActive 检查 refresh token 是否存在，过期
	IsRefreshTokenActive(ctx context.Context, userId string, jti string) (bool, error)

	// IsBlacklisted 检查 token 是否在黑名单中
	IsBlacklisted(ctx context.Context, jti string) (bool, error)

	// StoreRefreshToken 存储刷新令牌，用于轮换机制
	StoreRefreshToken(ctx context.Context, userID string, jti string, ttl time.Duration) error

	// DeleteRefreshToken 删除刷新令牌 (登出或轮换时使用)
	DeleteRefreshToken(ctx context.Context, userID string, jti string) error
}
