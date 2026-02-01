package config

import (
	"log"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	// 启用外键约束
	DB, err = gorm.Open(sqlite.Open("data/oinote.db?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		log.Fatal("无法连接数据库")
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.ChannelMember{},
		&models.Note{},
		&models.Attachment{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
}
