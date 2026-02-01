package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/MiXiaoAi/oinote/backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// cleanupEmptyDirs 递归删除空文件夹
func cleanupEmptyDirs(dirPath string) {
	// 防止删除uploads根目录
	if dirPath == "./uploads" || dirPath == "uploads" {
		return
	}

	// 检查目录是否为空
	isEmpty, err := isDirEmpty(dirPath)
	if err != nil {
		return // 如果无法检查，跳过删除
	}

	if isEmpty {
		// 尝试删除空目录
		os.Remove(dirPath)
		// 递归检查父目录
		parent := filepath.Dir(dirPath)
		if parent != dirPath && parent != "." {
			cleanupEmptyDirs(parent)
		}
	}
}

// isDirEmpty 检查目录是否为空
func isDirEmpty(dirPath string) (bool, error) {
	file, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = file.Readdirnames(1)
	if err == nil {
		return false, nil // 目录不为空
	}
	if err.Error() == "EOF" {
		return true, nil // 目录为空
	}
	return false, err // 其他错误
}

// cleanupEmptyDirsSafe 安全地递归删除空文件夹
func cleanupEmptyDirsSafe(dirPath string) {
	// 防止删除重要目录
	protectedDirs := []string{
		"./uploads",
		"uploads",
		"./uploads/channels",
		"./uploads/notes",
		"./uploads/temp",
	}

	for _, protected := range protectedDirs {
		if dirPath == protected {
			return
		}
	}

	// 检查目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return
	}

	// 检查目录是否为空
	isEmpty, err := isDirEmpty(dirPath)
	if err != nil || !isEmpty {
		return
	}

	// 尝试删除空目录
	if err := os.Remove(dirPath); err == nil {
		// 递归检查父目录，但限制递归深度
		parent := filepath.Dir(dirPath)
		if parent != dirPath && parent != "." && len(filepath.SplitList(parent)) < 10 {
			cleanupEmptyDirsSafe(parent)
		}
	}
}

type ChannelHandler struct {
	DB   *gorm.DB
	Hub  *websocket.Hub
}

func NewChannelHandler(db *gorm.DB, hub *websocket.Hub) *ChannelHandler {
	return &ChannelHandler{DB: db, Hub: hub}
}

func (h *ChannelHandler) CreateChannel(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	var channel models.Channel
	if err := c.BodyParser(&channel); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	channel.OwnerID = userId

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	h.DB.Model(&models.Channel{}).Order("id").Pluck("id", &existingIDs)

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
	channel.ID = nextAvailableID

	// 安全检查：如果复用了ID，确保该ID对应的旧数据已被完全清除
	// 检查是否有残留的成员记录
	var memberCount int64
	h.DB.Model(&models.ChannelMember{}).Where("channel_id = ?", nextAvailableID).Count(&memberCount)
	if memberCount > 0 {
		// 有残留的成员记录，需要清理
		h.DB.Exec("DELETE FROM channel_members WHERE channel_id = ?", nextAvailableID)
	}

	// 检查是否有残留的消息记录
	var messageCount int64
	h.DB.Model(&models.ChannelMessage{}).Where("channel_id = ?", nextAvailableID).Count(&messageCount)
	if messageCount > 0 {
		// 有残留的消息记录，需要清理
		h.DB.Exec("DELETE FROM channel_messages WHERE channel_id = ?", nextAvailableID)
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := h.DB.Exec("INSERT INTO channels (id, created_at, updated_at, name, description, owner_id, is_public, theme_color, tags) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?)",
		channel.ID, channel.Name, channel.Description, channel.OwnerID, channel.IsPublic, channel.ThemeColor, channel.Tags)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建频道失败: " + result.Error.Error()})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingMemberIDs []uint
	h.DB.Model(&models.ChannelMember{}).Order("id").Pluck("id", &existingMemberIDs)

	nextMemberID := uint(1)
	for _, id := range existingMemberIDs {
		if id == nextMemberID {
			nextMemberID++
		} else {
			// 找到间隙
			break
		}
	}

	// 自动加入所有者为 Owner 角色
	member := models.ChannelMember{
		ID:        nextMemberID,
		ChannelID: channel.ID,
		UserID:    userId,
		Role:      models.RoleOwner,
		Status:    models.MemberStatusActive,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	h.DB.Exec("INSERT INTO channel_members (id, channel_id, user_id, role, status, joined_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		member.ID, member.ChannelID, member.UserID, member.Role, member.Status)

	// 重新加载完整的频道信息，包括 Owner
	h.DB.Preload("Owner").First(&channel, channel.ID)

	// 广播频道创建消息
	h.Hub.BroadcastMessage("channel", "create", channel)

	return c.JSON(channel)
}

func (h *ChannelHandler) GetPublicChannels(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	var channels []models.Channel
	h.DB.Preload("Owner").Where("is_public = ?", true).Find(&channels)

	// 如果用户未登录，直接返回频道列表
	if userId == nil {
		return c.JSON(channels)
	}

	// 查询用户在这些频道中的成员记录
	channelIDs := make([]uint, len(channels))
	for i, ch := range channels {
		channelIDs[i] = ch.ID
	}

	var memberships []models.ChannelMember
	h.DB.Where("channel_id IN ? AND user_id = ?", channelIDs, userId).Find(&memberships)

	// 构建频道ID到成员状态的映射
	membershipMap := make(map[uint]models.ChannelMember)
	for _, m := range memberships {
		membershipMap[m.ChannelID] = m
	}

	// 将成员状态直接添加到频道对象中
	type PublicChannel struct {
		ID         uint       `json:"id"`
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
		Name       string     `json:"name"`
		Description string    `json:"description"`
		OwnerID    uint       `json:"owner_id"`
		IsPublic   bool       `json:"is_public"`
		ThemeColor string     `json:"theme_color"`
		Tags       string     `json:"tags"`
		Owner      models.User `json:"owner"`
		// 成员状态字段
		IsMember   bool   `json:"is_member"`
		IsPending  bool   `json:"is_pending"`
		IsInvited  bool   `json:"is_invited"`
		MemberRole string `json:"member_role"`
	}

	result := make([]PublicChannel, len(channels))
	for i, ch := range channels {
		result[i].ID = ch.ID
		result[i].CreatedAt = ch.CreatedAt
		result[i].UpdatedAt = ch.UpdatedAt
		result[i].Name = ch.Name
		result[i].Description = ch.Description
		result[i].OwnerID = ch.OwnerID
		result[i].IsPublic = ch.IsPublic
		result[i].ThemeColor = ch.ThemeColor
		result[i].Tags = ch.Tags
		result[i].Owner = ch.Owner

		// 填充成员状态
		if m, exists := membershipMap[ch.ID]; exists {
			result[i].IsMember = m.Status == models.MemberStatusActive
			result[i].IsPending = m.Status == models.MemberStatusPending
			result[i].IsInvited = m.Status == models.MemberStatusInvited
			result[i].MemberRole = m.Role
		}
	}

	return c.JSON(result)
}

func (h *ChannelHandler) GetUserChannels(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	var channels []models.Channel
	h.DB.Preload("Owner").Joins("JOIN channel_members ON channel_members.channel_id = channels.id").
		Where("channel_members.user_id = ? AND channel_members.status = ?", userId, models.MemberStatusActive).
		Find(&channels)
	return c.JSON(channels)
}

// GetChannel 获取单个频道及其成员信息
func (h *ChannelHandler) GetChannel(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	channelId := c.Params("id")

	var channel models.Channel
	if err := h.DB.First(&channel, "id = ?", channelId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在"})
	}

	// 如果频道是公开的
	if channel.IsPublic {
		// 如果用户已登录且是频道成员，显示完整成员列表
		if userId != nil {
			var membership models.ChannelMember
			err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", channelId, userId, models.MemberStatusActive).
				First(&membership).Error

			if err == nil {
				// 用户是成员，显示完整成员列表
				var members []models.ChannelMember
				h.DB.Preload("User").Where("channel_id = ? AND status = ?", channelId, models.MemberStatusActive).Find(&members)
				return c.JSON(fiber.Map{
					"channel": channel,
					"members": members,
				})
			}
		}
		// 访客或非成员访问，不显示成员列表
		return c.JSON(fiber.Map{
			"channel": channel,
			"members": nil,
		})
	}

	// 私有频道需要确认当前用户是该频道的成员
	if userId == nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
	}

	var membership models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", channelId, userId, models.MemberStatusActive).
		First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
	}

	// 查询成员列表并预加载用户信息
	var members []models.ChannelMember
	h.DB.Preload("User").Where("channel_id = ? AND status = ?", channelId, models.MemberStatusActive).Find(&members)

	return c.JSON(fiber.Map{
		"channel": channel,
		"members": members,
	})
}

func (h *ChannelHandler) GetChannelMessages(c *fiber.Ctx) error {
	channelId := c.Params("id")

	var channel models.Channel
	if err := h.DB.Preload("Owner").First(&channel, "id = ?", channelId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在"})
	}

	// 公开频道允许访客查看消息
	if !channel.IsPublic {
		userId := c.Locals("userId")
		if userId == nil {
			return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
		}

		var membership models.ChannelMember
		if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", channelId, userId, models.MemberStatusActive).
			First(&membership).Error; err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
		}
	}

	var messages []models.ChannelMessage
	h.DB.Preload("User").Preload("Attachment").
		Where("channel_id = ?", channelId).
		Order("created_at ASC").
		Find(&messages)

	return c.JSON(messages)
}

func (h *ChannelHandler) CreateChannelMessage(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")

	var membership models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", channelId, userId, models.MemberStatusActive).
		First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
	}

	type Input struct {
		Content      string `json:"content"`
		AttachmentID *uint  `json:"attachment_id"`
	}

	var input Input
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if input.Content == "" && input.AttachmentID == nil {
		return c.Status(400).JSON(fiber.Map{"error": "消息内容不能为空"})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	h.DB.Model(&models.ChannelMessage{}).Order("id").Pluck("id", &existingIDs)

	nextAvailableID := uint(1)
	for _, id := range existingIDs {
		if id == nextAvailableID {
			nextAvailableID++
		} else {
			// 找到间隙
			break
		}
	}

	message := models.ChannelMessage{
		ID:           nextAvailableID,
		ChannelID:    membership.ChannelID,
		UserID:       userId,
		Content:      input.Content,
		AttachmentID: input.AttachmentID,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := h.DB.Exec("INSERT INTO channel_messages (id, created_at, updated_at, channel_id, user_id, content, attachment_id, is_highlighted) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, 0)",
		message.ID, message.ChannelID, message.UserID, message.Content, message.AttachmentID)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "发送消息失败"})
	}

	h.DB.Preload("User").Preload("Attachment").First(&message, message.ID)

	// 广播新消息到所有客户端
	h.Hub.BroadcastMessage("message", "create", message)

	return c.JSON(message)
}

func (h *ChannelHandler) InviteUser(c *fiber.Ctx) error {
	// 简单实现邀请逻辑
	type InviteInput struct {
		ChannelID uint   `json:"channel_id"`
		Username  string `json:"username"`
	}
	var input InviteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	var user models.User
	if err := h.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "用户不存在"})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingMemberIDs []uint
	h.DB.Model(&models.ChannelMember{}).Order("id").Pluck("id", &existingMemberIDs)

	nextMemberID := uint(1)
	for _, id := range existingMemberIDs {
		if id == nextMemberID {
			nextMemberID++
		} else {
			// 找到间隙
			break
		}
	}

	member := models.ChannelMember{
		ID:        nextMemberID,
		ChannelID: input.ChannelID,
		UserID:    user.ID,
		Role:      models.RoleMember,
		Status:    models.MemberStatusInvited,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	if err := h.DB.Exec("INSERT INTO channel_members (id, channel_id, user_id, role, status, joined_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		member.ID, member.ChannelID, member.UserID, member.Role, member.Status).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "该用户已在频道中或已受邀"})
	}

	return c.JSON(fiber.Map{"message": "邀请已发送"})
}

func (h *ChannelHandler) RemoveMember(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")
	targetUserIdStr := c.Params("userId")

	targetUserId64, err := strconv.ParseUint(targetUserIdStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "用户参数无效"})
	}
	targetUserId := uint(targetUserId64)

	var currentMember models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ? AND (role = ? OR role = ?)",
		channelId, userId, models.MemberStatusActive, models.RoleOwner, models.RoleAdmin).First(&currentMember).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权操作"})
	}

	var targetMember models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?",
		channelId, targetUserId, models.MemberStatusActive).First(&targetMember).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "成员不存在"})
	}

	if targetMember.Role == models.RoleOwner {
		return c.Status(403).JSON(fiber.Map{"error": "不能移出频道所有者"})
	}

	if targetMember.UserID == userId {
		return c.Status(400).JSON(fiber.Map{"error": "不能移出自己"})
	}

	if err := h.DB.Delete(&targetMember).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "移出失败"})
	}

	return c.JSON(fiber.Map{"message": "已移出成员"})
}

func (h *ChannelHandler) UpdateChannel(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")

	var channel models.Channel
	// Check if channel exists and user is owner
	if err := h.DB.Where("id = ? AND owner_id = ?", channelId, userId).First(&channel).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在或无权修改"})
	}

	var input models.Channel
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	// Allow updating name, description, is_public and tags
	channel.Name = input.Name
	channel.Description = input.Description
	channel.IsPublic = input.IsPublic
	channel.Tags = input.Tags

	h.DB.Save(&channel)
	// 重新加载完整的频道信息，包括 Owner
	h.DB.Preload("Owner").First(&channel, channel.ID)

	// 广播频道更新消息
	h.Hub.BroadcastMessage("channel", "update", channel)

	return c.JSON(channel)
}

func (h *ChannelHandler) DeleteChannel(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")

	var channel models.Channel
	// Check if channel exists and user is owner
	if err := h.DB.Where("id = ? AND owner_id = ?", channelId, userId).First(&channel).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在或无权删除"})
	}

	// Transaction to delete channel, members, notes and related data
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 先统计要删除的成员数量（用于日志）
		var memberCount int64
		tx.Model(&models.ChannelMember{}).Where("channel_id = ?", channelId).Count(&memberCount)

		// Delete channel members - 使用原生SQL确保彻底删除
		if err := tx.Exec("DELETE FROM channel_members WHERE channel_id = ?", channelId).Error; err != nil {
			return err
		}

		// Delete channel messages
		if err := tx.Exec("DELETE FROM channel_messages WHERE channel_id = ?", channelId).Error; err != nil {
			return err
		}

		// Delete attachments
		if err := tx.Exec("DELETE FROM attachments WHERE channel_id = ?", channelId).Error; err != nil {
			return err
		}

		// Delete notes
		if err := tx.Exec("DELETE FROM notes WHERE channel_id = ?", channelId).Error; err != nil {
			return err
		}

		// Delete channel
		if err := tx.Exec("DELETE FROM channels WHERE id = ?", channelId).Error; err != nil {
			return err
		}

		// 验证删除是否成功
		var remainingCount int64
		tx.Model(&models.ChannelMember{}).Where("channel_id = ?", channelId).Count(&remainingCount)
		if remainingCount > 0 {
			return fmt.Errorf("删除后仍有 %d 条成员记录残留", remainingCount)
		}

		return nil
	})

	if err == nil {
		uploadDir := filepath.Join("./uploads", "channels", "channel_"+channelId)
		os.RemoveAll(uploadDir)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除频道失败"})
	}

	// 广播频道删除消息
	h.Hub.BroadcastMessage("channel", "delete", fiber.Map{
		"id": channelId,
	})

	return c.SendStatus(204)
}

// DeleteChannelMessage 删除消息（从数据库和文件系统中完全删除）
func (h *ChannelHandler) DeleteChannelMessage(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")
	messageId := c.Params("messageId")

	// 检查用户权限
	var membership models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ?", channelId, userId, models.MemberStatusActive).
		First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权访问该频道"})
	}

	var message models.ChannelMessage
	if err := h.DB.Where("id = ? AND channel_id = ?", messageId, channelId).First(&message).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "消息不存在"})
	}

	// 只有消息发送者可以删除自己的消息，或者管理员/所有者可以删除任何消息
	if message.UserID != userId && membership.Role != models.RoleAdmin && membership.Role != models.RoleOwner {
		return c.Status(403).JSON(fiber.Map{"error": "只能删除自己的消息"})
	}

	// 收集需要删除的附件信息
	var attachmentToDelete *models.Attachment

	// 开始事务删除消息和相关文件
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 如果消息有附件，标记为待删除
		if message.AttachmentID != nil {
			var attachment models.Attachment
			if err := tx.First(&attachment, *message.AttachmentID).Error; err == nil {
				attachmentToDelete = &attachment
				// 删除附件记录（文件稍后删除）
				tx.Delete(&attachment)
			}
		}

		// 删除消息记录
		if err := tx.Delete(&message).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "删除消息失败"})
	}

	// 广播消息删除到所有客户端
	h.Hub.BroadcastMessage("message", "delete", fiber.Map{
		"id":         message.ID,
		"channel_id": message.ChannelID,
	})

	// 延迟删除文件（等待浏览器释放引用）
	if attachmentToDelete != nil && attachmentToDelete.FilePath != "" {
		cwd, _ := os.Getwd()
		fullPath := filepath.Join(cwd, "data", attachmentToDelete.FilePath)
		absPath, _ := filepath.Abs(fullPath)

		go func() {
			time.Sleep(5 * time.Second)
			os.Remove(absPath)
		}()
	}

	return c.JSON(fiber.Map{"message": "消息已删除"})
}

// HighlightMessage 设置或取消精华消息
func (h *ChannelHandler) HighlightMessage(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	channelId := c.Params("id")
	messageId := c.Params("messageId")

	// 检查用户权限（只有管理员和所有者可以设置精华）
	var membership models.ChannelMember
	if err := h.DB.Where("channel_id = ? AND user_id = ? AND status = ? AND (role = ? OR role = ?)",
		channelId, userId, models.MemberStatusActive, models.RoleAdmin, models.RoleOwner).First(&membership).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权设置精华消息"})
	}

	var message models.ChannelMessage
	if err := h.DB.Where("id = ? AND channel_id = ?", messageId, channelId).First(&message).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "消息不存在"})
	}

	// 切换精华状态
	message.IsHighlighted = !message.IsHighlighted
	if err := h.DB.Save(&message).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "设置精华失败"})
	}

	// 广播精华状态更新到所有客户端
	h.Hub.BroadcastMessage("message", "highlight", fiber.Map{
		"id":              message.ID,
		"channel_id":      message.ChannelID,
		"is_highlighted":  message.IsHighlighted,
	})

	status := "取消精华"
	if message.IsHighlighted {
		status = "设为精华"
	}

	return c.JSON(fiber.Map{
		"message":        status,
		"is_highlighted": message.IsHighlighted,
	})
}
