package impl

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go_chat_v1_test/internal/auth"
	initconfig "go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/exceptions"
	"go_chat_v1_test/internal/middlewares"
	"go_chat_v1_test/internal/repository"
	"go_chat_v1_test/internal/services"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	userRepo      repository.UserRepository
	redisService  services.RedisServices
	accessSecret  string
	refreshSecret string
	accessTTL     string        // 例如: 15 * time.Minute
	refreshTTL    string        // 例如: 7 * 24 * time.Hour
	accessttl     time.Duration // 15m -> 15分钟
	refreshttl    time.Duration // 168h -> 7天
}

func NewAuthServiceImpl(userRepo repository.UserRepository, redisService services.RedisServices,
	jwtConfig initconfig.JwtConfig) services.AuthServices {
	accessttl, _ := time.ParseDuration(jwtConfig.AccessTTL)
	refreshttl, _ := time.ParseDuration(jwtConfig.RefreshTTL)
	return &AuthServiceImpl{
		userRepo:      userRepo,
		redisService:  redisService,
		accessSecret:  jwtConfig.AccessSecret,
		refreshSecret: jwtConfig.RefreshSecret,
		accessTTL:     jwtConfig.AccessTTL,
		refreshTTL:    jwtConfig.RefreshTTL,
		accessttl:     accessttl,
		refreshttl:    refreshttl,
	}
}

// Login 登录逻辑
func (s *AuthServiceImpl) Login(ctx context.Context, userEmail string, password string) (*auth.TokenPair, error) {
	// 1. 查询用户
	user, err := s.userRepo.FindByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, exceptions.ErrUserNotFoundInDB) {
			return nil, exceptions.ErrUserNotFoundInDB // 返回自定义错误
		}
		return nil, errors.New("invalid credentials")
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. 生成 Token 对
	tokens, err := auth.GenerateTokens(user.ID, s.accessSecret, s.refreshSecret, s.accessttl, s.refreshttl)
	if err != nil {
		return nil, err
	}

	// 4. 存储 Refresh Token 到 Redis
	claims, err := auth.ParseToken(tokens.RefreshToken, s.refreshSecret) // 提取 jti
	if err != nil {
		return nil, err
	}
	s.StoreRefreshToken(ctx, strconv.FormatUint(uint64(user.ID), 10), claims.ID, s.refreshttl)

	return tokens, nil
}

// Logout 登出逻辑
func (s *AuthServiceImpl) Logout(ctx context.Context, claims *auth.Claims) error {
	// 将 Access Token 的 jti 加入黑名单，直到它过期
	ttl := time.Until(claims.ExpiresAt.Time)
	return s.BlacklistToken(ctx, claims.ID, ttl)
}

// Refresh 刷新令牌
func (s *AuthServiceImpl) Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// 1. 验证 Refresh Token
	claims, err := auth.ParseToken(refreshToken, s.refreshSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 2. 检查 Redis 中是否存在 (确保是有效的)
	// 这里为了简洁，假设 Refresh Token 有效。完整实现应严格校验
	if claims != nil {
		var isActive, err = s.IsRefreshTokenActive(ctx, strconv.FormatUint(uint64(claims.UserID), 10), claims.ID)
		//log.Printf("%s %s", err.Error(), time.Now())
		middlewares.SafeLogError(err)
		if !isActive {
			return nil, exceptions.ErrUserAlreadyLogout
		}
		var isBlacked, _ = s.IsBlacklisted(ctx, claims.ID)
		if isBlacked {
			return nil, exceptions.ErrUserDisabled
		}
	}

	// 3. 生成新的 Token 对 (轮换)
	newTokens, err := auth.GenerateTokens(claims.UserID, s.accessSecret, s.refreshSecret, s.accessttl, s.refreshttl)
	if err != nil {
		return nil, err
	}

	// 4. (关键) 删除旧的 Refresh Token，存储新的 [citation:3]
	s.DeleteRefreshToken(ctx, strconv.FormatUint(uint64(claims.UserID), 10), claims.ID)
	newClaims, _ := auth.ParseToken(newTokens.RefreshToken, s.refreshSecret)
	s.StoreRefreshToken(ctx, strconv.FormatUint(uint64(claims.UserID), 10), newClaims.ID, s.refreshttl)

	return newTokens, nil
}

// BlacklistToken 将 token 的 jti 加入黑名单，有效期与 token 剩余时间一致
func (r *AuthServiceImpl) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	return r.redisService.Set(ctx, "blacklist:"+jti, "revoked", ttl)
}

// IsRefreshTokenActive 检查 refresh token 是否存在，过期
func (r *AuthServiceImpl) IsRefreshTokenActive(ctx context.Context, userId string, jti string) (bool, error) {
	val, err := r.redisService.Exists(ctx, "refresh:"+userId+":"+jti)
	if err != nil {
		return false, err
	}
	return val, nil
}

// IsBlacklisted 检查 token 是否在黑名单中
func (r *AuthServiceImpl) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	val, err := r.redisService.Exists(ctx, "blacklist:"+jti)
	if err != nil {
		return false, err
	}
	return val, nil
}

// StoreRefreshToken 存储刷新令牌，用于轮换机制
func (r *AuthServiceImpl) StoreRefreshToken(ctx context.Context, userID string, jti string, ttl time.Duration) error {
	return r.redisService.Set(ctx, "refresh:"+userID+":"+jti, "valid", ttl)
}

// DeleteRefreshToken 删除刷新令牌 (登出或轮换时使用)
func (r *AuthServiceImpl) DeleteRefreshToken(ctx context.Context, userID string, jti string) error {
	return r.redisService.Delete(ctx, "refresh:"+userID+":"+jti)
}
