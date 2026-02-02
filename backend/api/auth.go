package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MiXiaoAi/oinote/backend/internal/middleware"
	"github.com/MiXiaoAi/oinote/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if len(input.Username) < 3 || len(input.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "用户名或密码太短"})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)

	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Nickname: input.Username, // 默认昵称跟随用户名
	}

	// 查找第一个可用的空ID（填充ID间隙）
	var existingIDs []uint
	h.DB.Model(&models.User{}).Order("id").Pluck("id", &existingIDs)

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
	user.ID = nextAvailableID

	// 使用Raw SQL插入，确保使用指定的ID
	result := h.DB.Exec("INSERT INTO users (id, created_at, updated_at, username, password, nickname, avatar, bio) VALUES (?, datetime('now'), datetime('now'), ?, ?, ?, ?, ?)",
		user.ID, user.Username, user.Password, user.Nickname, user.Avatar, user.Bio)

	if result.Error != nil {
		return c.Status(400).JSON(fiber.Map{"error": "用户名已存在"})
	}

	return c.JSON(fiber.Map{"message": "注册成功"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	var user models.User
	if err := h.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "用户名或密码错误"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "用户名或密码错误"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"token": t,
		"user":  user,
	})
}

// GetStats 获取系统统计信息（仅管理员）
func (h *AuthHandler) GetStats(c *fiber.Ctx) error {
	var stats struct {
		UserCount    int64 `json:"user_count"`
		NoteCount    int64 `json:"note_count"`
		ChannelCount int64 `json:"channel_count"`
	}

	h.DB.Model(&models.User{}).Count(&stats.UserCount)
	h.DB.Model(&models.Note{}).Count(&stats.NoteCount)
	h.DB.Model(&models.Channel{}).Count(&stats.ChannelCount)

	return c.JSON(stats)
}

// GetAllUsers 获取所有用户列表（仅管理员）
func (h *AuthHandler) GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	h.DB.Order("created_at DESC").Find(&users)

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	return c.JSON(users)
}

// UpdateUserRole 更新用户角色（仅管理员）
func (h *AuthHandler) UpdateUserRole(c *fiber.Ctx) error {
	userId := c.Params("id")

	type UpdateRoleInput struct {
		Role string `json:"role"`
	}

	var input UpdateRoleInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	// 验证角色
	if input.Role != models.RoleAdmin && input.Role != models.RoleMember {
		return c.Status(400).JSON(fiber.Map{"error": "无效的角色"})
	}

	var user models.User
	if err := h.DB.First(&user, userId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "用户不存在"})
	}

	user.Role = input.Role
	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新失败"})
	}

	user.Password = "" // 清除密码字段
	return c.JSON(user)
}

// DeleteUser 删除用户（仅管理员）
func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	userId := c.Params("id")
	adminId := c.Locals("userId").(uint)

	// 检查是否尝试删除自己
	if adminId == h.convertToUint(userId) {
		return c.Status(400).JSON(fiber.Map{"error": "不能删除自己"})
	}

	var user models.User
	if err := h.DB.First(&user, userId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "用户不存在"})
	}

	// 删除用户相关的数据
	// 1. 删除频道成员关系
	h.DB.Exec("DELETE FROM channel_members WHERE user_id = ?", userId)

	// 2. 删除用户的笔记和相关附件
	// 先获取该用户的所有笔记
	var notes []models.Note
	h.DB.Where("owner_id = ?", userId).Find(&notes)

	// 删除每个笔记的附件和文件
	for _, note := range notes {
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
		h.DB.Exec("DELETE FROM attachments WHERE note_id = ?", note.ID)

		// 删除笔记目录
		noteDir := filepath.Join("./data/uploads/notes", fmt.Sprintf("note_%d", note.ID))
		os.RemoveAll(noteDir)
	}

	// 删除用户的笔记
	h.DB.Exec("DELETE FROM notes WHERE owner_id = ?", userId)

	// 3. 删除用户创建的频道（级联删除成员和消息）
	var channels []models.Channel
	h.DB.Where("owner_id = ?", userId).Find(&channels)

	for _, channel := range channels {
		// 删除频道目录
		channelDir := filepath.Join("./data/uploads/channels", fmt.Sprintf("channel_%d", channel.ID))
		os.RemoveAll(channelDir)
	}

	h.DB.Exec("DELETE FROM channels WHERE owner_id = ?", userId)

	// 4. 删除用户
	h.DB.Delete(&user)

	return c.JSON(fiber.Map{"message": "删除成功"})
}

// convertToUint 辅助函数，将字符串转换为 uint
func (h *AuthHandler) convertToUint(s string) uint {
	var result uint
	fmt.Sscanf(s, "%d", &result)
	return result
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	var user models.User
	h.DB.First(&user, userId)
	return c.JSON(user)
}

func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	userId := c.Locals("userId").(uint)

	var user models.User
	if err := h.DB.First(&user, userId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "用户不存在"})
	}

	type UpdateInput struct {
		Nickname *string `json:"nickname"`
		Avatar   *string `json:"avatar"`
		Bio      *string `json:"bio"`
	}

	var input UpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "输入数据无效"})
	}

	if input.Nickname != nil {
		user.Nickname = *input.Nickname
	}
	if input.Avatar != nil {
		user.Avatar = *input.Avatar
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "更新失败"})
	}

	return c.JSON(user)
}
