package i18n

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type I18n struct {
	viperMap map[string]*viper.Viper
	mu       sync.RWMutex
}

// NewI18n 创建国际化实例
func NewI18n(localesPath string) (*I18n, error) {
	i := &I18n{
		viperMap: make(map[string]*viper.Viper),
	}

	// 读取locales目录下的所有yaml文件
	files, err := filepath.Glob(filepath.Join(localesPath, "*.yaml"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// 从文件名获取语言代码（如 zh-CN.yaml -> zh-CN）
		fileName := filepath.Base(file)
		lang := strings.TrimSuffix(fileName, ".yaml")

		// 为每种语言创建一个viper实例
		v := viper.New()
		v.SetConfigFile(file)
		v.SetConfigType("yaml")

		// 读取配置文件
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取语言文件失败 %s: %w", file, err)
		}

		// 存储viper实例
		i.viperMap[lang] = v
	}

	return i, nil
}

// T 翻译消息（支持格式化参数）
func (i *I18n) T(lang, key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// 获取对应语言的viper实例
	v := i.getViper(lang)
	if v == nil {
		return key
	}

	// 获取消息模板
	msg := v.GetString(key)
	if msg == "" {
		return key // 找不到就返回key本身
	}

	// 如果有参数，格式化消息
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}

	return msg
}

// GetFieldName 获取字段的中文/英文名称
func (i *I18n) GetFieldName(lang, field string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	v := i.getViper(lang)
	if v == nil {
		return field
	}

	// 从fields映射中获取字段名
	fieldName := v.GetString("fields." + field)
	if fieldName == "" {
		return field
	}
	return fieldName
}

// GetErrorsName 获取字段的中文/英文名称
func (i *I18n) GetErrorsName(lang, field string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	v := i.getViper(lang)
	if v == nil {
		return field
	}

	// 从errors映射中获取字段名
	fieldName := v.GetString("errors." + field)
	if fieldName == "" {
		return field
	}
	return fieldName
}

func (i *I18n) getValidatorTargetName(field string) string {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 使用json标签作为字段名
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
	return field

}

// GetValidationMessage 获取验证错误消息
func (i *I18n) GetValidationMessage(lang, theFieldName string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	v := i.getViper(lang)
	if v == nil {
		return theFieldName
	}

	// 从validation映射中获取消息模板
	msg := v.GetString("validation." + i.getValidatorTargetName(theFieldName))
	if msg == "" {
		return theFieldName
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}

	return msg
}

// GetErrorsMessage 获取错误消息
func (i *I18n) GetErrorsMessage(lang string, theFieldName string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	v := i.getViper(lang)
	if v == nil {
		return theFieldName
	}

	// 从Errors映射中获取消息模板
	msg := v.GetString("errors." + theFieldName)
	if msg == "" {
		return theFieldName
	}

	return msg
}

// GetAllKeys 获取所有消息键（调试用）
func (i *I18n) GetAllKeys(lang string) []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if v, ok := i.viperMap[lang]; ok {
		return v.AllKeys()
	}
	return nil
}

// getViper 获取对应语言的viper实例
func (i *I18n) getViper(lang string) *viper.Viper {
	v, ok := i.viperMap[lang]
	if !ok {
		// 如果语言不存在，默认用中文
		v = i.viperMap["zh-CN"]
	}
	return v
}

// 全局单例
var (
	globalI18n *I18n
	once       sync.Once
)

// Init 初始化全局国际化（在main中调用）
func Init(localesPath string) error {
	var err error
	once.Do(func() {
		globalI18n, err = NewI18n(localesPath)
	})
	return err
}

// T 全局翻译函数
func T(lang string, key string, args ...interface{}) string {
	if globalI18n == nil {
		return key
	}
	return globalI18n.T(lang, key, args...)
}

func GetErrorsMessage(lang string, field string) string {
	if globalI18n == nil {
		return field
	}
	return globalI18n.GetErrorsMessage(lang, field)
}

func GetFieldName(lang string, field string) string {
	if globalI18n == nil {
		return field
	}
	return globalI18n.GetFieldName(lang, field)
}

func GetErrorsName(lang string, field string) string {
	if globalI18n == nil {
		return field
	}
	return globalI18n.GetErrorsName(lang, field)
}

func getDefaultI18n(lang string) *viper.Viper {
	if globalI18n == nil {
		return nil
	}
	return globalI18n.getViper(lang)
}

// GetValidationMessage 全局获取验证消息
func GetValidationMessage(lang string, rule string, args ...interface{}) string {
	if globalI18n == nil {
		return rule
	}
	return globalI18n.GetValidationMessage(lang, rule, args...)
}
