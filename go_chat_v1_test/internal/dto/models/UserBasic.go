package models

import (
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name          string     `gorm:"size:100"`
	Password      string     `gorm:"size:100"`
	Phone         string     `gorm:"size:20"`
	Email         string     `gorm:"size:100"`
	ClientIP      string     `gorm:"column:client_ip;size:50"`
	IdentityID    string     `gorm:"column:identity_id;size:50"`
	ClientPort    string     `gorm:"column:client_port;size:10"`
	LoginTime     *time.Time `gorm:"column:login_time"`
	HeartBeatTime *time.Time `gorm:"column:heart_beat_time"`
	LogoutTime    *time.Time `gorm:"column:logout_time"`
	IsLogout      bool       `gorm:"column:is_logout;default:false"`
	DeviceInfo    string     `gorm:"column:device_info;size:255"`
	IsDelete      bool
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

type ChatRoom struct {
	gorm.Model
	RoomName        string `gorm:"column:room_name;size:100"`
	RoomOwnerUserID string `gorm:"column:room_owner_user_id;size:100"`
	IsDelete        bool
}

func (table *ChatRoom) TableName() string {
	return "chat_room"
}
