//go:build wireinject
// +build wireinject

package container

import (
	"go_chat_v1_test/internal/chat"
	"go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/controllers"
	"go_chat_v1_test/internal/repository"
	"go_chat_v1_test/internal/services/impl"

	"github.com/google/wire"
)

var ControllerSet = wire.NewSet(
	controllers.NewUserController,
	controllers.NewAuthController,
	controllers.NewHealthController,
	controllers.NewCommonController,
	controllers.NewWebSocketController,
)

var ChatSet = wire.NewSet(
	chat.NewDefaultClientFactory,
	chat.NewDefaultRoomFactory,
	chat.NewWebSocketHandler,
)

// InitializeContainer 注入点
func InitializeContainer(configPath string) (*Container, func(), error) {
	wire.Build(
		// ==================== 配置相关 ====================
		ProvideConfig,
		// 提取所有需要的配置字段
		wire.FieldsOf(new(*config.Config),
			"App",           //总工程配置
			"Database",      // 数据库配置
			"Redis",         // Redis 配置
			"JwtInfo",       // JWT 配置
			"Log",           // 日志配置
			"KafkaInfo",     // Kafka 配置
			"WebSocketInfo", //WebSocket配置
		),

		// ==================== 基础组件 ====================
		// 基础Provider： 它们接收 *config.Config 并返回具体依赖
		ProvideLog,           // 返回全局log配置
		ProvideDB,            // 返回 *gorm.DB
		ProvideRedisService,  // 返回 *redis.Client
		ProvideKafkaProducer, // 返回 *kgo.Client only for Producer
		ProvideKafkaConsumer, // 返回 *kgo.Client only for Consumer
		ProvideShardedHub,    // 返回ShardHub
		// ==================== Repository 层 ====================
		// Repository Provider
		ProvideUserRepository,
		repository.NewUserRepositoryImpl, // 它需要 *gorm.DB，wire会自动注入 provideDB 的结果

		// ==================== Service 层 ====================
		// Service Providers
		impl.NewUserServiceImpl,
		impl.NewRedisServiceImpl,
		impl.NewKafkaServiceImpl,
		impl.NewCommonServiceImpl, // 假设它依赖 RedisService 或 *redis.Client
		impl.NewAuthServiceImpl,   // 它依赖 UserRepository, *redis.Client, 和 JWT配置

		// ==================== Controller 层 ====================
		// 使用 Set 统一管理，所有 Controller 只需在 controllers/wire.go 中维护
		ControllerSet,
		ChatSet,

		// ==================== 构造 Container ====================
		wire.Struct(new(Container), "*"),
	)
	// 这个返回值是给wire生成代码用的占位符，实际调用时不会执行到这里
	return &Container{}, nil, nil
}
