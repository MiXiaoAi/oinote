package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/MiXiaoAi/oinote/backend/config"
	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

func UploadFile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "上传失败"})
	}

	uploadType := c.FormValue("type") // "note" or "channel"

	// 限制检查
	if uploadType == "note" {
		if file.Size > 100*1024*1024 { // 100MB
			return c.Status(400).JSON(fiber.Map{"error": "笔记附件不能超过100MB"})
		}
	}
	// 频道文件无限制 (Server Config BodyLimit 仍生效)

	// 生成人类可读的文件名: 原文件名_用户ID_当前日期
	ext := filepath.Ext(file.Filename)
	baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])
	dateStr := time.Now().Format("20060102")
	newFileName := fmt.Sprintf("%s_%d_%s%s", baseName, userID, dateStr, ext)
	savePath := fmt.Sprintf("./data/uploads/%s", newFileName)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "保存文件失败"})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	config.DB.Model(&models.Attachment{}).Order("id").Pluck("id", &existingIDs)

	nextAvailableID := uint(1)
	for _, id := range existingIDs {
		if id == nextAvailableID {
			nextAvailableID++
		} else {
			// 找到间隙
			break
		}
	}

	// 记录到数据库
	attachment := models.Attachment{
		ID:         nextAvailableID,
		FileName:   file.Filename,
		FilePath:   "/uploads/" + newFileName,
		FileSize:   file.Size,
		FileType:   uploadType,
		UploaderID: userID,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := config.DB.Exec("INSERT INTO attachments (id, created_at, updated_at, file_name, file_path, file_size, file_type, uploader_id, channel_id, note_id) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, NULL, NULL)",
		attachment.ID, attachment.FileName, attachment.FilePath, attachment.FileSize, attachment.FileType, attachment.UploaderID)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "保存附件记录失败"})
	}

	return c.JSON(fiber.Map{
		"url": attachment.FilePath,
		"id":  attachment.ID,
	})
}
