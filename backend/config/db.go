package config

import (
	"log"
	"os"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	// 确保 data 目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	var err error
	// 启用外键约束
	DB, err = gorm.Open(sqlite.Open("data/oinote.db?_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.ChannelMember{},
		&models.Note{},
		&models.Attachment{},
		&models.ChannelMessage{},
		&models.AIConfig{},
	)
	if err != nil {
		return err
	}

	// 创建默认 admin 用户（如果用户表为空）
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminUser := models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Nickname: "管理员",
			Role:     "admin",
		}

		if err := DB.Create(&adminUser).Error; err != nil {
			return err
		}
		log.Println("已创建默认管理员用户: admin / admin")
	}
	return nil
}
