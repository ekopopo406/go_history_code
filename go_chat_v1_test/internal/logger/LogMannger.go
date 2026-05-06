package logger

import (
	"go_chat_v1_test/internal/config"
	"os"
	"path/filepath"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Manager struct {
	App    *zap.Logger
	DB     *zap.Logger
	Redis  *zap.Logger
	Kafka  *zap.Logger
	Access *zap.Logger
	Ws     *zap.Logger
	Error  *zap.Logger
}

// NewLogManager 创建日志管理器
func NewLogManager(cfg *config.LogConfig) (*Manager, error) {
	if cfg == nil {
		cfg = &config.LogConfig{Level: "info"}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建各个 logger
	return &Manager{
		App:    createLogger("app", cfg.AppLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		DB:     createLogger("db", cfg.DBLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		Redis:  createLogger("redis", cfg.RedisLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		Kafka:  createLogger("kafka", cfg.KafkaLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		Access: createLogger("access", cfg.AccessLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		Ws:     createLogger("Ws", cfg.WsLogPath, cfg, encoderConfig, parseLevel(cfg.Level)),
		Error:  createLogger("error", cfg.ErrorLogPath, cfg, encoderConfig, zapcore.ErrorLevel),
	}, nil
}

// 内部统一创建 logger 的函数
func createLogger(name, path string, cfg *config.LogConfig, encCfg zapcore.EncoderConfig, level zapcore.Level) *zap.Logger {
	writer := getWriter(path, cfg)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		writer,
		level,
	)

	return zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("logger", name)), // 给每条日志自动加上 logger 名称
	)
}

func getWriter(path string, cfg *config.LogConfig) zapcore.WriteSyncer {
	if path == "" {
		return zapcore.AddSync(os.Stdout)
	}

	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	lumberJackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync 优雅关闭所有 logger
func (m *Manager) Sync() error {
	var errs []error
	for _, l := range []*zap.Logger{m.App, m.DB, m.Redis, m.Kafka, m.Access, m.Error} {
		if l != nil {
			if err := l.Sync(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs[0] // 返回第一个错误即可
	}
	return nil
}

func (m *Manager) KafkaLogger() kgo.Logger {
	if m.Kafka == nil {
		return kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)
	}
	kl := m.Kafka.WithOptions(zap.AddCallerSkip(1))
	//kl := m.Kafka.WithOptions(zap.IncreaseLevel(zapcore.WarnLevel))
	return kzap.New(kl)
}
