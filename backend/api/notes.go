package handlers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/MiXiaoAi/oinote/backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type NoteHandler struct {
	DB  *gorm.DB
	Hub *websocket.Hub
}

func NewNoteHandler(db *gorm.DB, hub *websocket.Hub) *NoteHandler {
	return &NoteHandler{DB: db, Hub: hub}
}

func (h *NoteHandler) CreateNote(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	var note models.Note
	if err := c.BodyParser(&note); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	note.OwnerID = userId

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	h.DB.Model(&models.Note{}).Order("id").Pluck("id", &existingIDs)

	nextAvailableID := uint(1)
	for _, id := range existingIDs {
		if id == nextAvailableID {
			nextAvailableID++
		} else {
			// 找到间隙
			break
		}
	}

	// 手动设置ID以填充间隙
	note.ID = nextAvailableID

	// 使用Raw SQL插入，确保使用指定的ID
	result := h.DB.Exec("INSERT INTO notes (id, created_at, updated_at, title, content, channel_id, owner_id, is_public, tags) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?)",
		note.ID, note.Title, note.Content, note.ChannelID, note.OwnerID, note.IsPublic, note.Tags)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建笔记失败: " + result.Error.Error()})
	}

	// 重新加载笔记信息，包括所有者
	h.DB.Preload("Owner").First(&note, note.ID)

	// 广播笔记创建消息
	h.Hub.BroadcastMessage("note", "create", note)

	return c.JSON(note)
}

func (h *NoteHandler) GetNotes(c *fiber.Ctx) error {
	channelId := c.QueryInt("channel_id", 0)
	userId := c.Locals("userId")
	hasUserId := userId != nil

	var notes []models.Note

	if channelId != 0 {
		// 获取频道下的公开笔记
		query := h.DB.Where("channel_id = ? AND is_public = ?", channelId, true)
		query.Preload("Owner").Find(&notes)

		// 如果用户已登录，还返回用户自己的笔记（无论是否公开）
		if hasUserId {
			var userNotes []models.Note
			h.DB.Where("channel_id = ? AND owner_id = ?", channelId, userId.(uint)).Preload("Owner").Find(&userNotes)

			// 合并并去重
			noteMap := make(map[uint]models.Note)
			for _, n := range notes {
				noteMap[n.ID] = n
			}
			for _, n := range userNotes {
				noteMap[n.ID] = n
			}

			notes = make([]models.Note, 0, len(noteMap))
			for _, n := range noteMap {
				notes = append(notes, n)
			}
		}
	} else {
		// 未认证用户只能看公开笔记
		if !hasUserId {
			h.DB.Where("is_public = ?", true).Preload("Owner").Find(&notes)
		} else {
			// 认证用户看自己的所有笔记
			query := h.DB.Where("owner_id = ?", userId.(uint))
			query.Where("channel_id IS NULL")
			query.Preload("Owner").Find(&notes)
		}
	}

	return c.JSON(notes)
}

func (h *NoteHandler) GetPublicNotes(c *fiber.Ctx) error {
	var notes []models.Note
	h.DB.Where("is_public = ?", true).Preload("Owner").Find(&notes)
	return c.JSON(notes)
}

func (h *NoteHandler) SearchNotes(c *fiber.Ctx) error {
	queryStr := c.Query("q")
	if queryStr == "" {
		return c.JSON([]models.Note{})
	}

	var notes []models.Note
	
	// 获取用户ID（可能为nil，因为使用了OptionalAuth）
	userIdInterface := c.Locals("userId")
	
	if userIdInterface != nil {
		// 用户已登录：搜索用户自己的笔记或公开的笔记
		userId := userIdInterface.(uint)
		h.DB.Where("(owner_id = ? OR is_public = ?) AND (title LIKE ? OR content LIKE ? OR tags LIKE ?)", 
			userId, true, "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%").
			Preload("Owner").Find(&notes)
	} else {
		// 用户未登录：只搜索公开的笔记
		h.DB.Where("is_public = ? AND (title LIKE ? OR content LIKE ? OR tags LIKE ?)", 
			true, "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%").
			Preload("Owner").Find(&notes)
	}
	
	return c.JSON(notes)
}

func (h *NoteHandler) UpdateNote(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	noteId := c.Params("id")

	var note models.Note
	if err := h.DB.First(&note, noteId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "笔记不存在"})
	}

	// 权限检查：个人笔记只能作者编辑，频道笔记所有成员都能编辑
	canEdit := false
	if note.ChannelID != nil {
		// 频道笔记：检查是否是频道成员
		var membership models.ChannelMember
		if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", 
			*note.ChannelID, userId, models.MemberStatusActive).First(&membership).Error; err == nil {
			canEdit = true
		}
	} else {
		// 个人笔记：只有作者可以编辑
		canEdit = (note.OwnerID == userId)
	}

	if !canEdit {
		return c.Status(403).JSON(fiber.Map{"error": "笔记不存在或无权修改"})
	}

	type UpdateNoteInput struct {
		Title       *string  `json:"title"`
		Content     *string  `json:"content"`
		IsPublic    *bool    `json:"is_public"`
		Tags        *string  `json:"tags"`
		LineSpacing *float64 `json:"line_spacing"`
	}

	var input UpdateNoteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if input.Title != nil {
		note.Title = *input.Title
	}
	if input.Content != nil {
		note.Content = *input.Content
	}
	if input.IsPublic != nil {
		note.IsPublic = *input.IsPublic
	}
	if input.Tags != nil {
		note.Tags = *input.Tags
	}
	if input.LineSpacing != nil {
		note.LineSpacing = *input.LineSpacing
	}

	if err := h.DB.Save(&note).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新笔记失败"})
	}

	// 清理不再使用的附件
	if input.Content != nil {
		// 获取该笔记的所有附件
		var attachments []models.Attachment
		h.DB.Where("note_id = ?", note.ID).Find(&attachments)

		// 从新内容中提取所有文件路径
		newContent := *input.Content
		usedFilePaths := make(map[string]bool)

		// 使用正则表达式提取所有图片src和文件链接
		imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
		imgMatches := imgRegex.FindAllStringSubmatch(newContent, -1)
		for _, match := range imgMatches {
			if len(match) > 1 {
				filePath := match[1]
				if idx := strings.Index(filePath, "?"); idx != -1 {
					filePath = filePath[:idx]
				}
				if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
					if idx := strings.Index(filePath, "/uploads/"); idx != -1 {
						filePath = filePath[idx:]
					}
				}
				usedFilePaths[filePath] = true
			}
		}

		// 匹配链接
		linkRegex := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["']`)
		linkMatches := linkRegex.FindAllStringSubmatch(newContent, -1)
		for _, match := range linkMatches {
			if len(match) > 1 {
				filePath := match[1]
				if idx := strings.Index(filePath, "?"); idx != -1 {
					filePath = filePath[:idx]
				}
				if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
					if idx := strings.Index(filePath, "/uploads/"); idx != -1 {
						filePath = filePath[idx:]
					}
				}
				usedFilePaths[filePath] = true
			}
		}

		// 收集待删除的附件
		var toDelete []models.Attachment
		for _, attachment := range attachments {
			if !usedFilePaths[attachment.FilePath] {
				toDelete = append(toDelete, attachment)
			}
		}

		// 延迟删除文件（等待浏览器释放引用）
		if len(toDelete) > 0 {
			// 立即删除数据库记录
			for _, attachment := range toDelete {
				h.DB.Delete(&attachment)
			}

			// 5秒后删除文件
			go func() {
				time.Sleep(5 * time.Second)
				for _, attachment := range toDelete {
					if attachment.FilePath != "" {
						cwd, _ := os.Getwd()
						fullPath := filepath.Join(cwd, "data", attachment.FilePath)
						absPath, _ := filepath.Abs(fullPath)
						os.Remove(absPath)
					}
				}
			}()
		}
	}

	// 重新加载完整的笔记信息，包括 Owner
	h.DB.Preload("Owner").First(&note, note.ID)

	// 广播笔记更新消息
	h.Hub.BroadcastMessage("note", "update", note)

	return c.JSON(note)
}

func (h *NoteHandler) GetNote(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	noteId := c.Params("id")

	var note models.Note
	if err := h.DB.Preload("Owner").First(&note, "id = ?", noteId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "笔记不存在"})
	}

	// 公开笔记允许访客访问，否则只允许所有者访问
	if !note.IsPublic {
		if userId == nil {
			return c.Status(403).JSON(fiber.Map{"error": "无权访问该笔记"})
		}
		if note.OwnerID != userId.(uint) {
			return c.Status(403).JSON(fiber.Map{"error": "无权访问该笔记"})
		}
	}

	return c.JSON(note)
}

func (h *NoteHandler) DeleteNote(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	noteId := c.Params("id")

	var note models.Note
	// Check if note exists and user owns it
	if err := h.DB.Where("id = ? AND owner_id = ?", noteId, userId).First(&note).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "笔记不存在或无权删除"})
	}

	// 获取该笔记的所有附件
	var attachments []models.Attachment
	h.DB.Where("note_id = ?", note.ID).Find(&attachments)

	// 删除所有附件文件
	for _, attachment := range attachments {
		if attachment.FilePath != "" {
			fullPath := filepath.Join("./data", attachment.FilePath)
			os.Remove(fullPath)
		}
	}

	// 删除附件记录
	h.DB.Where("note_id = ?", note.ID).Delete(&models.Attachment{})

	// 删除笔记目录
	noteDir := filepath.Join("./data/uploads/notes", "note_"+noteId)
	os.RemoveAll(noteDir)

	if err := h.DB.Delete(&note).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除笔记失败"})
	}

	// 广播笔记删除消息
	h.Hub.BroadcastMessage("note", "delete", fiber.Map{
		"id": noteId,
	})

	return c.SendStatus(204)
}

// GetAllNotes 管理员获取所有笔记
func (h *NoteHandler) GetAllNotes(c *fiber.Ctx) error {
	var notes []models.Note
	result := h.DB.Preload("Owner").Order("created_at DESC").Find(&notes)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "获取笔记失败"})
	}

	// 清除密码字段
	for i := range notes {
		if notes[i].Owner.Password != "" {
			notes[i].Owner.Password = ""
		}
	}

	return c.JSON(notes)
}

// AdminDeleteNote 管理员删除笔记
func (h *NoteHandler) AdminDeleteNote(c *fiber.Ctx) error {
	noteId := c.Params("id")

	var note models.Note
	if err := h.DB.First(&note, noteId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "笔记不存在"})
	}

	// 获取该笔记的所有附件
	var attachments []models.Attachment
	h.DB.Where("note_id = ?", note.ID).Find(&attachments)

	// 删除所有附件文件
	for _, attachment := range attachments {
		if attachment.FilePath != "" {
			fullPath := filepath.Join("./data", attachment.FilePath)
			os.Remove(fullPath)
		}
	}

	// 删除附件记录
	h.DB.Where("note_id = ?", note.ID).Delete(&models.Attachment{})

	// 删除笔记目录
	noteDir := filepath.Join("./data/uploads/notes", "note_"+noteId)
	os.RemoveAll(noteDir)

	if err := h.DB.Delete(&note).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除笔记失败"})
	}

	return c.JSON(fiber.Map{"message": "删除成功"})
}
