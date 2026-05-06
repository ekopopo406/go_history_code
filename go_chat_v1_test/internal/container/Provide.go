package container

import (
	"context"
	"fmt"
	"go_chat_v1_test/internal/chat"
	"go_chat_v1_test/internal/config"
	"go_chat_v1_test/internal/dto/models"
	"go_chat_v1_test/internal/logger"
	"go_chat_v1_test/internal/repository"
	"go_chat_v1_test/internal/types"
	apptypes "go_chat_v1_test/internal/types"
	"log"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	zapgorm "moul.io/zapgorm2"

	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
)

func ProvideShardedHub(cfg *config.WebSocketConfig, roomFactory chat.RoomFactory, clientFactory chat.ClientFactory, zapLogger *logger.Manager) (*chat.ShardedHub, func(), error) {
	log.Println("🚀 启动 WebSocket Hub goroutine")
	shardedHub := chat.NewShardedHub(cfg.ShardCount)

	for _, hub := range shardedHub.GetAllHubs() {
		hub.SetRoomFactory(roomFactory)
		hub.SetClientFactory(clientFactory)
	}
	time.Sleep(50 * time.Millisecond)

	zapLogger.Ws.Info("WebSocket Hub 已启动在 :8080")
	cleanup := func() {
		zapLogger.Ws.Info("[WebSocket Hub] 开始优雅关闭...")
		zapLogger.Ws.Info("[WebSocket Hub] 优雅关闭完成")
	}
	return shardedHub, cleanup, nil
}

func ProvideLog(cfg *config.LogConfig) (*logger.Manager, func(), error) {

	manager, err := logger.NewLogManager(cfg)

	if err != nil {
		log.Printf("初始化 LogManager 失败: %v", err)
		panic("logger Manager init FAILED!")
	}
	cleanup := func() {
		if manager.App != nil {
			manager.App.Info("开始执行 Logger Sync...")
		} else {
			log.Println("开始执行 Logger Sync...")
		}

		if syncErr := manager.Sync(); syncErr != nil {
			if manager.Error != nil {
				manager.Error.Error("Logger Sync 失败", zap.Error(syncErr))
			} else {
				log.Printf("Logger Sync 失败: %v", syncErr)
			}
		}

		if manager.App != nil {
			manager.App.Info("Logger Sync 完成")
		}
	}
	return manager, cleanup, nil
}

// provideDB 需要 Config 和 Logger
func ProvideDB(cfgApp *config.AppConfig, cfgDB *config.DatabaseConfig, cfgLog *config.LogConfig, zapLogger *logger.Manager) (types.MainDB, func(), error) {
	// 用 zap 创建 GORM 专用的 logger
	gormZapLogger := zapgorm.New(zapLogger.DB.With(zap.String("component", "gorm")))

	gormLogLevel := config.GetLogLevel(cfgLog.ShowGormSQL, cfgApp.Mode)

	gormLogger := gormZapLogger.LogMode(gormLogLevel)

	db, err := gorm.Open(mysql.Open(cfgDB.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}
	cleanup := func() {
		log.Println("正在关闭 Database...")
		sqlDB, _ := db.DB()
		if err := sqlDB.Close(); err != nil {
			log.Printf("Database 关闭错误: %v", err)
		}
	}
	db.AutoMigrate(&models.ChatRoom{})
	return types.MainDB{Db: db}, cleanup, nil
}

// provideRedisService 需要 Config
func ProvideRedisService(cfg *config.RedisConfig, zapLogger *logger.Manager) (types.MainRedis, func(), error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		// 根据需要：panic、返回 error、log.Fatal 等
		zapLogger.Redis.Panic("Redis 启动错误: %v", zap.Error(err))
	}

	cleanup := func() {
		log.Println("正在关闭 Redis...")
		if err := rdb.Close(); err != nil {
			zapLogger.Redis.Info("Redis 关闭错误: %v", zap.Error(err))
		}
	}
	return types.MainRedis{RedisClient: rdb}, cleanup, nil
}

// provideKafka 需要 Config
// func ProvideKafka(cfg *config.KafkaConfig, zapLogger *logger.Manager) (types.MainKafka, func(), error) {
// 	kafkaOpts := []kgo.Opt{
// 		kgo.SeedBrokers(cfg.KafkaBrokers), // []string 从配置读
// 		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
// 		kgo.AllowAutoTopicCreation(), // 开发方便，生产慎用
// 		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
// 		kgo.RequiredAcks(kgo.AllISRAcks()),       // 安全
// 		kgo.ProducerBatchMaxBytes(1_000_000),     // 1MB
// 		kgo.ProducerLinger(5 * time.Millisecond), // 小延迟高吞吐
// 	}

// 	client, err := kgo.NewClient(kafkaOpts...)
// 	if err != nil {
// 		zapLogger.Kafka.Info("kafka 连接失败: %v", zap.Error(err))
// 		return types.MainKafka{}, nil, err
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 根据你的业务调整超时
// 	defer cancel()

// 	// 关键步骤1：先 Flush 所有缓冲的消息（Producer）
// 	if err := client.Flush(ctx); err != nil {
// 		if ctx.Err() != nil {
// 			zapLogger.Kafka.Warn("Kafka Flush 超时，部分消息可能丢失", zap.Error(err))
// 		} else {
// 			zapLogger.Kafka.Error("Kafka Flush 失败", zap.Error(err))
// 		}
// 	} else {
// 		log.Println("Kafka Producer 所有缓冲消息已 Flush")
// 	}

// 	cleanup := func() {
// 		log.Println("正在关闭 kafka...")
// 		client.Close()
// 	}

// 	return types.MainKafka{Client: client}, cleanup, nil
// }

// 只用于 Producer 的 Client
func ProvideKafkaProducer(cfg *config.KafkaConfig, zapLogger *logger.Manager) (types.MainKafkaProducer, func(), error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers),
		kgo.WithLogger(zapLogger.KafkaLogger()),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchMaxBytes(cfg.ProducerBatchMaxBytes),
		kgo.ProducerLinger(5 * time.Millisecond),
		kgo.AllowAutoTopicCreation(), // 开发方便，生产关闭
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		zapLogger.Kafka.Error("Kafka Producer 创建失败", zap.Error(err))
		return types.MainKafkaProducer{}, nil, err
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		if err := client.Flush(ctx); err != nil {
			zapLogger.Kafka.Warn("Kafka Producer Flush 超时或失败", zap.Error(err))
		}
		client.Close()
		zapLogger.Kafka.Info("Kafka Producer 已优雅关闭")
	}

	return types.MainKafkaProducer{ProducerClient: client}, cleanup, nil
}

// 只用于 Consumer 的 Client
func ProvideKafkaConsumer(cfg *config.KafkaConfig, zapLogger *logger.Manager) (types.MainKafkaConsumer, func(), error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers),
		kgo.ConsumerGroup(cfg.ConsumerGroupID),
		kgo.ConsumeTopics(cfg.Topics),
		kgo.WithLogger(zapLogger.KafkaLogger()),
		// Consumer 专用选项：
		// kgo.ConsumerGroupRebalanceStrategy(...),
		// kgo.DisableAutoCommit() 如果手动 commit
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		zapLogger.Kafka.Error("Kafka Consumer 创建失败", zap.Error(err))
		return types.MainKafkaConsumer{}, nil, err
	}

	cleanup := func() {
		// Consumer 关闭时通常会自动 commit（如果启用 auto commit）
		// 如果你禁用 auto commit，这里可以手动 CommitUncommittedOffsets
		client.Close()
		zapLogger.Kafka.Info("Kafka Consumer 已关闭")
	}

	return types.MainKafkaConsumer{ConsumerClient: client}, cleanup, nil
}

func ProvideConfig(configPath string) *config.Config {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	return cfg
}

func ProvideUserRepository(mainDb apptypes.MainDB) repository.Repository[models.UserBasic] {
	return repository.NewGORMRepository[models.UserBasic](mainDb)
}
