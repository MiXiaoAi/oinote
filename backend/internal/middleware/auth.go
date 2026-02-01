package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
)

var JwtSecret = []byte("oinote_secret_key_123456")

func AuthRequired(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "未授权"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "无效的令牌"})
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("userId", uint(claims["user_id"].(float64)))
	c.Locals("username", claims["username"].(string))

	return c.Next()
}

// OptionalAuth 可选认证中间件，如果有token则解析，没有则继续
func OptionalAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Next() // token无效时继续，不返回错误
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("userId", uint(claims["user_id"].(float64)))
	c.Locals("username", claims["username"].(string))

	return c.Next()
}
