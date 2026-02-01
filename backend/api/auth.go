package handlers

import (
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
