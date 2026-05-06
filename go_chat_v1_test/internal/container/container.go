package container

import (
	"go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/controllers"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/repository"
	"go_chat_v1_test/internal/services"
	apptypes "go_chat_v1_test/internal/types"
)

type Container struct {
	Config              *config.Config
	DB                  apptypes.MainDB
	logger              *logger.Manager
	UserRepo            repository.UserRepository
	UserService         services.UserServices
	UserController      *controllers.UserController
	AuthController      *controllers.AuthController
	HealthController    *controllers.HealthController
	CommonController    *controllers.CommonController
	WebSocketController *controllers.WebSocketController
	RedisService        services.RedisServices
	AuthService         services.AuthServices
	KafkaService        services.KafkaServices
	CommonService       services.CommonServices
}

// NewContainer 现在由 wire 生成
func NewContainer(configPath string) (*Container, func(), error) {
	return InitializeContainer(configPath)
}
