package handlers

import (
	"github.com/MiXiaoAi/oinote/backend/config"
	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// CreateNote 创建笔记
func CreateNote(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input models.Note
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入无效"})
	}

	input.OwnerID = userID

	// 如果是频道笔记，检查成员权限
	if input.ChannelID != nil {
		var member models.ChannelMember
		if err := config.DB.Where("channel_id = ? AND user_id = ? AND status = ?",
			*input.ChannelID, userID, models.MemberStatusActive).First(&member).Error; err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "你不是该频道成员"})
		}
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	config.DB.Model(&models.Note{}).Order("id").Pluck("id", &existingIDs)

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
	input.ID = nextAvailableID

	// 使用Raw SQL插入，确保使用指定的ID
	result := config.DB.Exec("INSERT INTO notes (id, created_at, updated_at, title, content, channel_id, owner_id, is_public, tags) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?)",
		input.ID, input.Title, input.Content, input.ChannelID, input.OwnerID, input.IsPublic, input.Tags)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建笔记失败: " + result.Error.Error()})
	}

	return c.JSON(input)
}

// SearchNotes 搜索笔记
func SearchNotes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	keyword := c.Query("q")

	var notes []models.Note
	// 逻辑：搜索(我的笔记 OR 我所在频道的笔记 OR 公开笔记) AND 包含关键字
	// 这里简化实现，分为两步：先搜内容，再在代码层或复杂SQL层过滤权限。
	// 为了性能和安全性，使用 Gorm 的 Scope 或 Raw SQL 更好。此处演示逻辑：

	// 查找用户加入的所有频道ID
	var memberChannelIDs []uint
	config.DB.Model(&models.ChannelMember{}).Where("user_id = ? AND status = ?", userID, models.MemberStatusActive).Pluck("channel_id", &memberChannelIDs)

	db := config.DB.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	// 权限过滤条件：(公开) OR (是作者) OR (属于用户所在的频道)
	db = db.Where("is_public = ? OR owner_id = ? OR channel_id IN ?", true, userID, memberChannelIDs)

	db.Find(&notes)
	return c.JSON(notes)
}

// GetNoteDetail 获取详情
func GetNoteDetail(c *fiber.Ctx) error {
	// 需要补全详细的权限检查，类似 SearchNotes
	return c.Status(200).JSON(fiber.Map{"message": "TODO"})
}
