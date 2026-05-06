package impl

import (
	"context"
	"crypto/rand"
	"go_chat_v1_test/internal/services"
	"math/big"
	"time"
)

type CommonService struct {
	redisService services.RedisServices
}

func NewCommonServiceImpl(redisService services.RedisServices) services.CommonServices {
	return &CommonService{redisService: redisService}
}

// SendEmailVerifyCode implements services.CommonServices.
func (c *CommonService) SendEmailVerifyCode(ctx context.Context, input services.SendEmailVerifyCodeInput) int {

	return 0
}

// SendPhoneCode implements services.CommonServices.
func (c *CommonService) SendPhoneCode(ctx context.Context, input services.SendPhoneCodeInput) int {
	// 生成 16 字节的随机字节切片
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		panic(err)
	}
	var codeStr = int(n.Int64())
	c.redisService.Set(ctx, "PHONE:"+input.PhoneNumber, codeStr, 30*time.Second)
	return codeStr
}
