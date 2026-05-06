package main

import (
	"context"
	"flag"
	"go_chat_v1_test/internal/container"
	"go_chat_v1_test/internal/i18n"
	routers "go_chat_v1_test/internal/router"
	"go_chat_v1_test/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// func printConfig(cfg *config.Config) {
// 	fmt.Println("========== 应用配置 ==========")
// 	fmt.Printf("应用名称: %s\n", cfg.App.Name)
// 	fmt.Printf("应用版本: %s\n", cfg.App.Version)
// 	fmt.Printf("监听端口: %s\n", cfg.App.Port)
// 	fmt.Printf("运行模式: %s\n", cfg.App.Mode)

// 	fmt.Println("\n========== 数据库配置 ==========")
// 	fmt.Printf("数据库类型: %s\n", cfg.Database.Driver)
// 	fmt.Printf("数据库地址: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
// 	fmt.Printf("数据库名称: %s\n", cfg.Database.DBName)
// 	fmt.Printf("最大连接数: %d\n", cfg.Database.MaxConnections)

// 	fmt.Println("\n========== Redis配置 ==========")
// 	fmt.Printf("Redis地址: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
// 	fmt.Printf("Redis DB: %d\n", cfg.Redis.DB)
// 	fmt.Printf("连接池大小: %d\n", cfg.Redis.PoolSize)

// 	fmt.Println("\n========== 日志配置 ==========")
// 	fmt.Printf("日志级别: %s\n", cfg.Log.Level)

// 	fmt.Println("\n========== JWT配置 ==========")
// 	fmt.Printf("AccessSecret: %s\n", cfg.JwtInfo.AccessSecret)
// 	fmt.Printf("RefreshSecret: %s\n", cfg.JwtInfo.RefreshSecret)
// 	fmt.Printf("AccessTTL: %s\n", cfg.JwtInfo.AccessTTL)
// 	fmt.Printf("RefreshTTL: %s\n", cfg.JwtInfo.RefreshTTL)

// 	fmt.Println("\n========== Kafka配置 ==========")
// 	fmt.Printf("KafkaBrokers: %s\n", cfg.KafkaInfo.KafkaBrokers)

// 	// fmt.Println("\n========== 功能开关 ==========")
// 	// fmt.Printf("启用Metrics: %v\n", cfg.Features.EnableMetrics)
// 	// fmt.Printf("启用Swagger: %v\n", cfg.Features.EnableSwagger)
// 	// fmt.Printf("允许的域名: %v\n", cfg.Features.AllowedOrigins)
// }

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	var configPath string
	flag.StringVar(&configPath, "application", "./", "配置文件路径")
	flag.Parse()

	container, cleanup, err := container.InitializeContainer(configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	if err := i18n.Init("./internal/locales"); err != nil {
		panic("初始化国际化失败: " + err.Error())
	}

	router := routers.SetupRouter(container)
	// 启动 HTTP 服务
	srv := server.NewServer(container.Config.App.Port, router, logger)

	go func() {
		logger.Info("服务启动", zap.String("port", container.Config.App.Port))
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("服务启动失败", zap.Error(err))
		}
	}()

	//优雅关闭
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("优雅关闭失败", zap.Error(err))
	}
	logger.Info("服务已安全退出")
}
