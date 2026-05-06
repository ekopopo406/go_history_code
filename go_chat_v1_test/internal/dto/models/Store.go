package models

import "gorm.io/gorm"

type Model interface {
	gorm.Model  // 类型必须嵌入或拥有 gorm.Model 的字段
	GetID() any // 返回 ID（可以是 uint / string / int 等）
}
