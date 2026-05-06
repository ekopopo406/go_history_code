package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go_chat_v1_test/internal/i18n"

	"github.com/go-playground/validator/v10"
)

// Validator 验证器结构体
type Validator struct {
	validate *validator.Validate
}

// NewValidator 创建验证器实例
func NewValidator() *Validator {
	v := &Validator{
		validate: validator.New(),
	}

	// 注册自定义验证规则
	v.registerCustomValidations()

	return v
}

// registerCustomValidations 注册自定义验证规则
func (v *Validator) registerCustomValidations() {
	// 手机号验证
	v.validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// 简单的手机号验证：1开头的11位数字
		return len(phone) == 11 && phone[0] == '1'
	})
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(obj interface{}) error {
	return v.validate.Struct(obj)
}

// FormatValidationErrors 格式化验证错误（使用i18n）
func (v *Validator) FormatValidationErrors(err error, lang string) string {
	result := make(map[string]string)
	var msg string

	// 类型断言为validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// result["error"] = err.Error()
		// return result
		msg = err.Error()
	}

	for _, e := range validationErrors {
		// 获取字段名（从json标签）
		field := e.Field()

		// 获取字段的中文/英文名称
		fieldName := i18n.GetFieldName(lang, strings.ToLower(field))

		// 根据验证规则获取对应的错误消息
		if e.Param() != "" {
			msg = i18n.GetValidationMessage(lang, e.Tag(), fieldName, e.Param())
		} else {
			msg = i18n.GetValidationMessage(lang, e.Tag(), fieldName)
		}
		result[field] = msg
	}

	return msg
}

// ParseAndValidate 解析JSON并验证（一站式函数）
func (v *Validator) ParseAndValidate(r *http.Request, obj interface{}, lang string) error {
	// 解析JSON
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}
	defer r.Body.Close()

	// 验证
	return v.ValidateStruct(obj)
}
