package handlers

import (
	"fmt"
	"time"

	"github.com/MiXiaoAi/oinote/backend/config"
	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

// CreateChannel 创建频道
func CreateChannel(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var input models.Channel
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入无效"})
	}

	input.OwnerID = userID

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	config.DB.Model(&models.Channel{}).Order("id").Pluck("id", &existingIDs)

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

	// 安全检查：如果复用了ID，确保该ID对应的旧数据已被完全清除
	// 检查是否有残留的成员记录
	var memberCount int64
	config.DB.Model(&models.ChannelMember{}).Where("channel_id = ?", nextAvailableID).Count(&memberCount)
	if memberCount > 0 {
		// 有残留的成员记录，需要清理
		config.DB.Exec("DELETE FROM channel_members WHERE channel_id = ?", nextAvailableID)
	}

	// 检查是否有残留的消息记录
	var messageCount int64
	config.DB.Model(&models.ChannelMessage{}).Where("channel_id = ?", nextAvailableID).Count(&messageCount)
	if messageCount > 0 {
		// 有残留的消息记录，需要清理
		config.DB.Exec("DELETE FROM channel_messages WHERE channel_id = ?", nextAvailableID)
	}

	// 使用Raw SQL插入，确保使用指定的ID
	result := config.DB.Exec("INSERT INTO channels (id, created_at, updated_at, name, description, owner_id, is_public, theme_color, tags) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?, ?)",
		input.ID, input.Name, input.Description, input.OwnerID, input.IsPublic, input.ThemeColor, input.Tags)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "创建失败: " + result.Error.Error()})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingMemberIDs []uint
	config.DB.Model(&models.ChannelMember{}).Order("id").Pluck("id", &existingMemberIDs)

	nextMemberID := uint(1)
	for _, id := range existingMemberIDs {
		if id == nextMemberID {
			nextMemberID++
		} else {
			// 找到间隙
			break
		}
	}

	// 自动将创建者加入为Owner
	member := models.ChannelMember{
		ID:        nextMemberID,
		ChannelID: input.ID,
		UserID:    userID,
		Role:      models.RoleOwner,
		Status:    models.MemberStatusActive,
		JoinedAt:  time.Now(),
	}

	// 使用Raw SQL插入，确保使用指定的ID
	config.DB.Exec("INSERT INTO channel_members (id, channel_id, user_id, role, status, joined_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		member.ID, member.ChannelID, member.UserID, member.Role, member.Status)

	return c.JSON(input)
}

// GetUserChannels 获取用户已加入的频道
func GetUserChannels(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	var members []models.ChannelMember
	// 只查询 Active 的频道
	config.DB.Preload("Channel").Where("user_id = ? AND status = ?", userID, models.MemberStatusActive).Find(&members)

	var channels []models.Channel
	for _, m := range members {
		channels = append(channels, m.Channel)
	}
	return c.JSON(channels)
}

// GetAllChannels 管理员获取所有频道
func GetAllChannels(c *fiber.Ctx) error {
	var channels []models.Channel
	config.DB.Preload("Owner").Find(&channels)

	// 为每个频道添加成员数
	for i := range channels {
		var memberCount int64
		config.DB.Model(&models.ChannelMember{}).Where("channel_id = ? AND status = ?", channels[i].ID, models.MemberStatusActive).Count(&memberCount)
		channels[i].MemberCount = int(memberCount)
	}

	return c.JSON(channels)
}

// AdminToggleChannelPublic 管理员切换频道公开/私密状态
func AdminToggleChannelPublic(c *fiber.Ctx) error {
	// 权限检查：只有管理员可以操作
	userId := c.Locals("userId")
	if userId == nil {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	var user models.User
	if err := config.DB.First(&user, userId).Error; err != nil || user.Role != models.RoleAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "无权操作"})
	}

	channelId := c.Params("id")

	type TogglePublicInput struct {
		IsPublic bool `json:"is_public"`
	}

	var input TogglePublicInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	var channel models.Channel
	if err := config.DB.First(&channel, channelId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在"})
	}

	channel.IsPublic = input.IsPublic
	if err := config.DB.Save(&channel).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新失败"})
	}

	// 重新加载完整的频道信息，包括 Owner
	config.DB.Preload("Owner").First(&channel, channel.ID)

	return c.JSON(channel)
}

// SearchPublicChannels 搜索公开频道
func SearchPublicChannels(c *fiber.Ctx) error {
	query := c.Query("q")
	var channels []models.Channel
	db := config.DB.Where("is_public = ?", true)
	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}
	db.Find(&channels)
	return c.JSON(channels)
}

// InviteUser 邀请用户 (Owner/Admin操作)
func InviteUser(c *fiber.Ctx) error {
	currentUserID := c.Locals("user_id").(uint)
	var input struct {
		ChannelID    uint `json:"channel_id"`
		TargetUserID uint `json:"target_user_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入无效"})
	}

	// 权限检查：当前用户必须是 Admin 或 Owner
	var currentUserMember models.ChannelMember
	if err := config.DB.Where("channel_id = ? AND user_id = ? AND (role = ? OR role = ?)",
		input.ChannelID, currentUserID, models.RoleOwner, models.RoleAdmin).First(&currentUserMember).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "无权邀请"})
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingMemberIDs []uint
	config.DB.Model(&models.ChannelMember{}).Order("id").Pluck("id", &existingMemberIDs)

	nextMemberID := uint(1)
	for _, id := range existingMemberIDs {
		if id == nextMemberID {
			nextMemberID++
		} else {
			// 找到间隙
			break
		}
	}

	// 创建邀请记录 (Status = invited)
	invite := models.ChannelMember{
		ID:        nextMemberID,
		ChannelID: input.ChannelID,
		UserID:    input.TargetUserID,
		Role:      models.RoleMember,
		Status:    models.MemberStatusInvited,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	config.DB.Exec("INSERT INTO channel_members (id, channel_id, user_id, role, status, joined_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		invite.ID, invite.ChannelID, invite.UserID, invite.Role, invite.Status)

	return c.JSON(fiber.Map{"message": "邀请已发送"})
}

// JoinChannelRequest 用户申请加入公开频道
func (h *ChannelHandler) JoinChannelRequest(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	channelIdStr := c.Params("id")
	
	// 解析频道ID
	var channelId uint
	_, err := fmt.Sscanf(channelIdStr, "%d", &channelId)
	if err != nil || channelId == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "无效的频道ID"})
	}

	// 检查频道是否公开
	var channel models.Channel
	if err := h.DB.First(&channel, channelId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "频道不存在"})
	}

	if !channel.IsPublic {
		return c.Status(403).JSON(fiber.Map{"error": "该频道为私有频道，无法直接申请"})
	}

	// 检查是否已经申请过
	var existingMember models.ChannelMember
	err = h.DB.Where("channel_id = ? AND user_id = ?", channelId, userID).First(&existingMember).Error
	if err == nil {
		// 已经存在记录
		if existingMember.Status == models.MemberStatusPending {
			return c.Status(400).JSON(fiber.Map{"error": "您已提交申请，请等待管理员审核"})
		} else if existingMember.Status == models.MemberStatusActive {
			return c.Status(400).JSON(fiber.Map{"error": "您已经是该频道的成员"})
		} else if existingMember.Status == models.MemberStatusInvited {
			// 如果之前是被邀请状态，现在用户主动申请，更新为待审核状态
			existingMember.Status = models.MemberStatusPending
			if err := h.DB.Save(&existingMember).Error; err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "申请失败，请稍后重试"})
			}
			return c.JSON(fiber.Map{"message": "申请已提交"})
		}
		return c.Status(400).JSON(fiber.Map{"error": "您已申请过该频道"})
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

	// 创建申请记录
	request := models.ChannelMember{
		ID:        nextMemberID,
		ChannelID: channelId,
		UserID:    userID,
		Role:      models.RoleMember,
		Status:    models.MemberStatusPending,
	}

	// 使用Raw SQL插入，确保使用指定的ID
	if err := h.DB.Exec("INSERT INTO channel_members (id, channel_id, user_id, role, status, joined_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		request.ID, request.ChannelID, request.UserID, request.Role, request.Status).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "申请失败，请稍后重试"})
	}

	return c.JSON(fiber.Map{"message": "申请已提交"})
}

// HandleMemberStatus 处理成员状态变更 (同意邀请 / 批准申请)
func (h *ChannelHandler) HandleMemberStatus(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	var input struct {
		MemberRecordID uint   `json:"member_record_id"`
		Action         string `json:"action"` // "accept_invite", "approve_request", "reject"
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).SendString("Invalid")
	}

	var memberRecord models.ChannelMember
	if err := h.DB.First(&memberRecord, input.MemberRecordID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "记录不存在"})
	}

	// 场景 1: 用户同意邀请
	if input.Action == "accept_invite" {
		if memberRecord.UserID != userID {
			return c.Status(403).SendString("非本人操作")
		}
		if memberRecord.Status != models.MemberStatusInvited {
			return c.Status(400).SendString("状态错误")
		}
		memberRecord.Status = models.MemberStatusActive
		memberRecord.JoinedAt = time.Now()
		if err := h.DB.Save(&memberRecord).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "操作失败"})
		}

		// 发送欢迎消息到频道
		go h.sendWelcomeMessage(memberRecord.ChannelID, memberRecord.UserID)
		
		// 广播频道更新事件，通知前端刷新成员状态
		var channel models.Channel
		h.DB.First(&channel, memberRecord.ChannelID)
		h.Hub.BroadcastMessage("channel", "update", channel)

		return c.JSON(fiber.Map{"message": "已加入频道"})
	}

	// 场景 2: 管理员批准申请
	if input.Action == "approve_request" {
		// 检查操作者是否有该频道的管理权限
		var adminCheck models.ChannelMember
		h.DB.Where("channel_id = ? AND user_id = ? AND (role = ? OR role = ?)",
			memberRecord.ChannelID, userID, models.RoleOwner, models.RoleAdmin).First(&adminCheck)
		if adminCheck.ID == 0 {
			return c.Status(403).SendString("无管理权限")
		}

		memberRecord.Status = models.MemberStatusActive
		memberRecord.Role = models.RoleMember
		memberRecord.JoinedAt = time.Now()
		if err := h.DB.Save(&memberRecord).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "操作失败"})
		}

		// 发送欢迎消息到频道
		go h.sendWelcomeMessage(memberRecord.ChannelID, memberRecord.UserID)
		
		// 广播频道更新事件，通知前端刷新成员状态
		var channel models.Channel
		h.DB.First(&channel, memberRecord.ChannelID)
		h.Hub.BroadcastMessage("channel", "update", channel)

		return c.JSON(fiber.Map{"message": "已批准加入"})
	}

	// 场景 3: 管理员拒绝申请
	if input.Action == "reject" {
		// 检查操作者是否有该频道的管理权限
		var adminCheck models.ChannelMember
		h.DB.Where("channel_id = ? AND user_id = ? AND (role = ? OR role = ?)",
			memberRecord.ChannelID, userID, models.RoleOwner, models.RoleAdmin).First(&adminCheck)
		if adminCheck.ID == 0 {
			return c.Status(403).SendString("无管理权限")
		}

		h.DB.Delete(&memberRecord)
		
		// 广播频道更新事件，通知前端刷新成员状态
		var channel models.Channel
		h.DB.First(&channel, memberRecord.ChannelID)
		h.Hub.BroadcastMessage("channel", "update", channel)
		
		return c.JSON(fiber.Map{"message": "已拒绝申请"})
	}

	return c.Status(400).SendString("未知操作")
}

// sendWelcomeMessage 发送欢迎消息到频道
func (h *ChannelHandler) sendWelcomeMessage(channelID uint, newUserID uint) {
	// 获取新用户信息
	var newUser models.User
	if err := h.DB.First(&newUser, newUserID).Error; err != nil {
		return
	}

	// 创建系统欢迎消息
	message := models.ChannelMessage{
		ChannelID: channelID,
		UserID:    0, // 0 表示系统消息
		Content:   fmt.Sprintf("🎉 欢迎 %s 加入频道！", newUser.Nickname),
	}

	if err := h.DB.Create(&message).Error; err == nil {
		h.DB.Preload("User").First(&message, message.ID)

		// 广播欢迎消息
		h.Hub.BroadcastMessage("message", "create", message)
	}
}

// GetPendingApprovals 获取当前用户的待处理消息（申请和邀请）
func (h *ChannelHandler) GetPendingApprovals(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)

	// 查找用户作为管理员或所有者的频道
	var managedChannels []models.ChannelMember
	result := h.DB.Where("user_id = ? AND status = ? AND (role = ? OR role = ?)",
		userID, models.MemberStatusActive, models.RoleOwner, models.RoleAdmin).
		Preload("Channel").
		Find(&managedChannels)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "查询失败"})
	}

	// 查找用户收到的邀请（用户是被邀请的一方）
	var userInvitations []models.ChannelMember
	result = h.DB.Preload("User").
		Preload("Channel").
		Preload("Channel.Owner").
		Where("user_id = ? AND status = ?", userID, models.MemberStatusInvited).
		Find(&userInvitations)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "查询失败"})
	}

	// 如果没有管理频道，只返回邀请
	if len(managedChannels) == 0 {
		return c.JSON(userInvitations)
	}

	// 获取这些频道的待审核申请
	channelIDs := make([]uint, len(managedChannels))
	for i, mc := range managedChannels {
		channelIDs[i] = mc.ChannelID
	}

	var pendingApprovals []models.ChannelMember
	result = h.DB.Preload("User").
		Preload("Channel").
		Where("channel_id IN ? AND status = ?", channelIDs, models.MemberStatusPending).
		Find(&pendingApprovals)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "查询失败"})
	}

	// 合并邀请和申请
	allMessages := append(userInvitations, pendingApprovals...)

	return c.JSON(allMessages)
}

// DeleteApproval 拒绝申请或删除邀请记录
func (h *ChannelHandler) DeleteApproval(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	approvalID := c.Params("id")

	// 查找申请或邀请记录
	var memberRecord models.ChannelMember
	if err := h.DB.First(&memberRecord, approvalID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "记录不存在"})
	}

	// 如果是邀请状态，允许被邀请人删除（拒绝邀请）
	if memberRecord.Status == models.MemberStatusInvited {
		if memberRecord.UserID != userID {
			return c.Status(403).JSON(fiber.Map{"error": "无权限操作"})
		}
		h.DB.Delete(&memberRecord)
		return c.JSON(fiber.Map{"message": "已拒绝邀请"})
	}

	// 如果是申请状态，需要管理员权限
	if memberRecord.Status == models.MemberStatusPending {
		// 检查操作者是否有该频道的管理权限
		var adminCheck models.ChannelMember
		h.DB.Where("channel_id = ? AND user_id = ? AND (role = ? OR role = ?)",
			memberRecord.ChannelID, userID, models.RoleOwner, models.RoleAdmin).First(&adminCheck)
		if adminCheck.ID == 0 {
			return c.Status(403).JSON(fiber.Map{"error": "无管理权限"})
		}
		// 删除申请记录
		h.DB.Delete(&memberRecord)
		return c.JSON(fiber.Map{"message": "已拒绝申请"})
	}

	return c.Status(400).JSON(fiber.Map{"error": "记录状态无效"})
}

// AcceptInvitation 接受邀请
func (h *ChannelHandler) AcceptInvitation(c *fiber.Ctx) error {
	userID := c.Locals("userId").(uint)
	invitationID := c.Params("id")

	// 查找邀请记录
	var memberRecord models.ChannelMember
	if err := h.DB.First(&memberRecord, invitationID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "记录不存在"})
	}

	// 检查是否是当前用户的邀请
	if memberRecord.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "无权限操作"})
	}

	// 检查状态是否为邀请
	if memberRecord.Status != models.MemberStatusInvited {
		return c.Status(400).JSON(fiber.Map{"error": "邀请状态无效"})
	}

	// 更新状态为活跃
	memberRecord.Status = models.MemberStatusActive
	memberRecord.Role = models.RoleMember
	memberRecord.JoinedAt = time.Now()

	if err := h.DB.Save(&memberRecord).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "接受邀请失败"})
	}

	// 发送系统消息通知频道成员
	h.sendJoinMessage(memberRecord.ChannelID, userID)
	
	// 广播频道更新事件，通知前端刷新成员状态
	var channel models.Channel
	h.DB.First(&channel, memberRecord.ChannelID)
	h.Hub.BroadcastMessage("channel", "update", channel)

	return c.JSON(fiber.Map{"message": "已接受邀请"})
}

// sendJoinMessage 发送加入消息到频道
func (h *ChannelHandler) sendJoinMessage(channelID uint, newUserID uint) {
	// 获取新用户信息
	var newUser models.User
	if err := h.DB.First(&newUser, newUserID).Error; err != nil {
		return
	}

	// 创建系统加入消息
	message := models.ChannelMessage{
		ChannelID: channelID,
		UserID:    0, // 0 表示系统消息
		Content:   fmt.Sprintf("🎉 %s 加入了频道！", newUser.Nickname),
	}

	if err := h.DB.Create(&message).Error; err == nil {
		h.DB.Preload("User").First(&message, message.ID)

		// 广播加入消息
		h.Hub.BroadcastMessage("message", "create", message)
	}
}
