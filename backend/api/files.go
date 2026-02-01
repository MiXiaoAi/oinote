package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type FileHandler struct {
	DB *gorm.DB
}

func NewFileHandler(db *gorm.DB) *FileHandler {
	return &FileHandler{DB: db}
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "上传失败"})
	}

	fileType := c.FormValue("type", "note_attachment")

	// 限制检查
	if fileType == "note_attachment" {
		if file.Size > 100*1024*1024 {
			return c.Status(400).JSON(fiber.Map{"error": "笔记附件不能超过100MB"})
		}
	}

	noteIdStr := c.FormValue("note_id")
	channelIdStr := c.FormValue("channel_id")

	var noteID *uint
	if noteIdStr != "" {
		if nid, err := strconv.ParseUint(noteIdStr, 10, 64); err == nil {
			id := uint(nid)
			noteID = &id
		}
	}

	var channelID *uint
	if channelIdStr != "" {
		if cid, err := strconv.ParseUint(channelIdStr, 10, 64); err == nil {
			id := uint(cid)
			channelID = &id
		}
	}

	// 确保目录存在并根据类型划分子目录，并为每个笔记/频道单独目录
	baseDir := "./data/uploads"
	subDir := ""

	switch fileType {
	case "avatar":
		subDir = "avatars"
	case "channel", "channel_file":
		contentType := file.Header.Get("Content-Type")
		channelFolder := "channels"
		if channelID != nil {
			channelFolder = filepath.Join("channels", fmt.Sprintf("channel_%d", *channelID))
		}
		if strings.HasPrefix(contentType, "image/") {
			subDir = filepath.Join(channelFolder, "images")
		} else if strings.HasPrefix(contentType, "video/") {
			subDir = filepath.Join(channelFolder, "videos")
		} else {
			subDir = filepath.Join(channelFolder, "files")
		}
	case "note", "note_attachment":
		noteFolder := "notes"
		if noteID != nil {
			noteFolder = filepath.Join("notes", fmt.Sprintf("note_%d", *noteID))
		}
		subDir = noteFolder
	default:
		subDir = "others"
	}

	uploadDir := filepath.Join(baseDir, subDir)
	os.MkdirAll(uploadDir, 0755)

	// 生成人类可读的文件名: 原文件名_用户ID_当前日期
	ext := filepath.Ext(file.Filename)
	baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])
	dateStr := time.Now().Format("20060102")
	newName := fmt.Sprintf("%s_%d_%s%s", baseName, userId, dateStr, ext)
	savePath := filepath.Join(uploadDir, newName)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "保存文件失败"})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	h.DB.Model(&models.Attachment{}).Order("id").Pluck("id", &existingIDs)

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
		FilePath:   "/" + filepath.ToSlash(filepath.Join("uploads", subDir, newName)),
		FileSize:   file.Size,
		FileType:   fileType,
		UploaderID: userId,
		ChannelID:  channelID,
		NoteID:     noteID,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := h.DB.Exec("INSERT INTO attachments (id, created_at, updated_at, file_name, file_path, file_size, file_type, uploader_id, channel_id, note_id) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?, ?)",
		attachment.ID, attachment.FileName, attachment.FilePath, attachment.FileSize, attachment.FileType, attachment.UploaderID, attachment.ChannelID, attachment.NoteID)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "保存附件记录失败"})
	}

	return c.JSON(attachment)
}
