package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	uploadType := c.FormValue("type") // "note", "channel", "avatar", or "channel_file"

	// 限制检查
	if uploadType == "note" || uploadType == "note_attachment" {
		if file.Size > 100*1024*1024 { // 100MB
			return c.Status(400).JSON(fiber.Map{"error": "笔记附件不能超过100MB"})
		}
	}
	if uploadType == "avatar" {
		if file.Size > 5*1024*1024 { // 5MB
			return c.Status(400).JSON(fiber.Map{"error": "头像不能超过5MB"})
		}
	}
	// 频道文件无限制 (Server Config BodyLimit 仍生效)

	var newFileName, savePath string
	ext := filepath.Ext(file.Filename)

	// 头像使用用户ID作为文件名，支持覆盖
	if uploadType == "avatar" {
		// 头像目录
		avatarDir := "./data/uploads/avatars"
		// 确保目录存在
		if err := os.MkdirAll(avatarDir, 0755); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "创建头像目录失败"})
		}

		newFileName = fmt.Sprintf("%d%s", userID, ext)
		savePath = filepath.Join(avatarDir, newFileName)

		// 删除旧头像（所有扩展名的旧头像）
		oldAvatarPattern := filepath.Join(avatarDir, fmt.Sprintf("%d.*", userID))
		matches, _ := filepath.Glob(oldAvatarPattern)
		for _, match := range matches {
			// 跳过新文件（如果文件名相同）
			if filepath.Base(match) == newFileName {
				continue
			}
			os.Remove(match)
		}
	} else if uploadType == "note" {
		// 笔记文件：保存到 notes/note_{noteID}/ 目录
		noteIDStr := c.FormValue("note_id")
		if noteIDStr == "" {
			return c.Status(400).JSON(fiber.Map{"error": "缺少笔记ID"})
		}

		// 人类可读的文件名
		baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])
		dateStr := time.Now().Format("20060102")
		newFileName = fmt.Sprintf("%s_%d_%s%s", baseName, userID, dateStr, ext)

		// 笔记目录：./data/uploads/notes/note_{noteID}/
		noteDir := filepath.Join("./data/uploads/notes", "note_"+noteIDStr)
		if err := os.MkdirAll(noteDir, 0755); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "创建笔记目录失败"})
		}
		savePath = filepath.Join(noteDir, newFileName)
	} else if uploadType == "channel" || uploadType == "channel_file" {
		// 频道文件：保存到 channels/channel_{channelID}/ 目录
		channelIDStr := c.FormValue("channel_id")
		if channelIDStr == "" {
			return c.Status(400).JSON(fiber.Map{"error": "缺少频道ID"})
		}

		// 人类可读的文件名
		baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])
		dateStr := time.Now().Format("20060102")
		newFileName = fmt.Sprintf("%s_%d_%s%s", baseName, userID, dateStr, ext)

		// 频道目录：./data/uploads/channels/channel_{channelID}/
		channelDir := filepath.Join("./data/uploads/channels", "channel_"+channelIDStr)
		if err := os.MkdirAll(channelDir, 0755); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "创建频道目录失败"})
		}
		savePath = filepath.Join(channelDir, newFileName)
	} else {
		// 其他类型（如 other）：保存到 others/ 目录
		baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])
		dateStr := time.Now().Format("20060102")
		newFileName = fmt.Sprintf("%s_%d_%s%s", baseName, userID, dateStr, ext)
		savePath = fmt.Sprintf("./data/uploads/others/%s", newFileName)
		os.MkdirAll(filepath.Dir(savePath), 0755)
	}

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
	// 根据文件类型确定文件路径
	var filePath string
	var noteID *uint
	var channelID *uint
	var fileType string

	if uploadType == "avatar" {
		fileType = "avatar"
		filePath = "/uploads/avatars/" + newFileName
	} else if uploadType == "note" || uploadType == "note_attachment" {
		fileType = "note"
		noteIDStr := c.FormValue("note_id")
		noteIDUint, _ := strconv.ParseUint(noteIDStr, 10, 32)
		nid := uint(noteIDUint)
		noteID = &nid
		filePath = "/uploads/notes/note_" + noteIDStr + "/" + newFileName
	} else if uploadType == "channel" || uploadType == "channel_file" {
		fileType = "channel"
		channelIDStr := c.FormValue("channel_id")
		channelIDUint, _ := strconv.ParseUint(channelIDStr, 10, 32)
		cid := uint(channelIDUint)
		channelID = &cid
		filePath = "/uploads/channels/channel_" + channelIDStr + "/" + newFileName
	} else {
		fileType = "other"
		filePath = "/uploads/others/" + newFileName
	}

	attachment := models.Attachment{
		ID:         nextAvailableID,
		FileName:   file.Filename,
		FilePath:   filePath,
		FileSize:   file.Size,
		FileType:   fileType,
		UploaderID: userID,
		NoteID:     noteID,
		ChannelID:  channelID,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := config.DB.Exec("INSERT INTO attachments (id, created_at, updated_at, file_name, file_path, file_size, file_type, uploader_id, channel_id, note_id) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?, ?)",
		attachment.ID, attachment.FileName, attachment.FilePath, attachment.FileSize, attachment.FileType, attachment.UploaderID, channelID, noteID)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "保存附件记录失败"})
	}

	return c.JSON(fiber.Map{
		"url": attachment.FilePath,
		"id":  attachment.ID,
	})
}
