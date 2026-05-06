package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

// AppConfig 应用配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Port    string `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxConnections  int    `mapstructure:"max_connections"`
	IdleConnections int    `mapstructure:"idle_connections"`
	SSLMode         string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"dbname"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JwtConfig struct {
	AccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
	RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTTL     string `mapstructure:"JWT_ACCESS_TTL"`  // 例如: 15 * time.Minute
	RefreshTTL    string `mapstructure:"JWT_REFRESH_TTL"` // 例如: 7 * 24 * time.Hour
}

// DSN 获取数据库连接字符串
func (c *DatabaseConfig) DSN() string {

	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.DBName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
	default:
		return ""
	}
}

// kafka 配置
type KafkaConfig struct {
	KafkaBrokers          string `mapstructure:"seedBorker"`
	Topics                string `mapstructure:"topics"`
	ConsumerGroupID       string `mapstructure:"consumerGroupID"`
	ProducerBatchMaxBytes int32  `mapstructure:"producerBatchMaxBytes"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`

	// 日志文件路径
	AppLogPath    string `mapstructure:"app_log_path"`
	DBLogPath     string `mapstructure:"db_log_path"`
	RedisLogPath  string `mapstructure:"redis_log_path"`
	KafkaLogPath  string `mapstructure:"kafka_log_path"`
	AccessLogPath string `mapstructure:"access_log_path"`
	ErrorLogPath  string `mapstructure:"error_log_path"`
	WsLogPath     string `mapstructure:"ws_log_path"`

	// 轮转配置
	MaxSize    int  `mapstructure:"max_size"`
	MaxBackups int  `mapstructure:"max_backups"`
	MaxAge     int  `mapstructure:"max_age"`
	Compress   bool `mapstructure:"compress"`

	// GORM 日志级别
	ShowGormSQL string `mapstructure:"zapGorm.level"`
}

type WebSocketConfig struct {
	ID             string        `yaml:"id"`
	ShardCount     int           `yaml:"shard_count"`
	WriteWait      time.Duration `yaml:"write_wait"`
	PongWait       time.Duration `yaml:"pong_wait"`
	PingPeriod     time.Duration `yaml:"ping_period"`
	MaxMessageSize int64         `yaml:"max_message_size"`
}

type MonitoringConfig struct {
	PrometheusPort string `yaml:"prometheus_port"`
}

// Config 总配置结构体
type Config struct {
	App           AppConfig       `mapstructure:"app"`
	Database      DatabaseConfig  `mapstructure:"database"`
	Redis         RedisConfig     `mapstructure:"redis"`
	Log           LogConfig       `mapstructure:"log"`
	JwtInfo       JwtConfig       `mapstructure:"jwt"`
	KafkaInfo     KafkaConfig     `mapstructure:"kafka"`
	WebSocketInfo WebSocketConfig `mapstructure:"websocket"`
	//ThirdParty ThirdPartyConfig `mapstructure:"third_party"`
	//Features   FeaturesConfig   `mapstructure:"features"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("application") // 配置文件名称(无扩展名)
	viper.SetConfigType("yaml")        // 如果配置文件没有扩展名，需要设置类型
	viper.AddConfigPath(configPath)    // 查找配置文件路径
	viper.AddConfigPath(".")           // 还可以在工作目录查找
	viper.AddConfigPath("./internal")  // 在config目录查找
	viper.AddConfigPath("../..")       // 在上级目录的config查找

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 打印配置文件路径（调试用）
	fmt.Println("使用配置文件:", viper.ConfigFileUsed())

	var config Config

	// 将配置解析到结构体
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	if config.WebSocketInfo.ShardCount == 0 {
		config.WebSocketInfo.ShardCount = 16
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name 不能为空")
	}

	if c.Database.Driver != "" {
		if c.Database.Host == "" {
			return fmt.Errorf("database.host 不能为空")
		}
		if c.Database.Port <= 0 {
			return fmt.Errorf("database.port 必须大于0")
		}
	}

	return nil
}

func GetLogLevel(levelStr string, appMode string) logger.LogLevel {
	switch levelStr {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	case "debug":
		return logger.Info // GORM 没有 Debug 级别，用 Info 代替
	}

	// 默认值：生产环境只打印错误，开发环境打印所有 SQL
	if appMode == "production" {
		return logger.Error
	}
	return logger.Info
}

// GetMode 获取应用模式
func (c *Config) GetMode() string {
	return c.App.Mode
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Mode == "production"
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Mode == "development"
}
