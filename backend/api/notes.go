package handlers

import (
	"os"
	"path/filepath"

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
	userId := c.Locals("userId").(uint)
	queryStr := c.Query("q")
	if queryStr == "" {
		return c.JSON([]models.Note{})
	}

	var notes []models.Note
	h.DB.Where("owner_id = ? AND (title LIKE ? OR content LIKE ?)", userId, "%"+queryStr+"%", "%"+queryStr+"%").
		Find(&notes)
	return c.JSON(notes)
}

func (h *NoteHandler) UpdateNote(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	noteId := c.Params("id")

	var note models.Note
	if err := h.DB.Where("id = ? AND owner_id = ?", noteId, userId).First(&note).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "笔记不存在或无权修改"})
	}

	type UpdateNoteInput struct {
		Title    *string `json:"title"`
		Content  *string `json:"content"`
		IsPublic *bool   `json:"is_public"`
		Tags     *string `json:"tags"`
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

	if err := h.DB.Save(&note).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新笔记失败"})
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

	h.DB.Where("note_id = ?", note.ID).Delete(&models.Attachment{})

	uploadDir := filepath.Join("./uploads", "notes", "note_"+noteId)
	os.RemoveAll(uploadDir)

	if err := h.DB.Delete(&note).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除笔记失败"})
	}

	// 广播笔记删除消息
	h.Hub.BroadcastMessage("note", "delete", fiber.Map{
		"id": noteId,
	})

	return c.SendStatus(204)
}
