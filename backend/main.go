package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	handlers "github.com/MiXiaoAi/oinote/backend/api"
	ws "github.com/MiXiaoAi/oinote/backend/internal/websocket"
	"github.com/MiXiaoAi/oinote/backend/internal/collab"
	"github.com/MiXiaoAi/oinote/backend/internal/middleware"
	"github.com/MiXiaoAi/oinote/backend/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// 初始化数据库
	if err := config.Connect(); err != nil {
		log.Fatal("无法连接数据库:", err)
	}

	db := config.DB

	// 初始化 WebSocket Hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// 初始化协同编辑服务器
	yjsServer := collab.NewYjsServer()

	// 初始化 Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 2 * 1024 * 1024 * 1024, // 2GB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,HEAD",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Range",
		ExposeHeaders:    "Content-Length,Accept-Ranges,Content-Range",
		AllowCredentials: false,
	}))
	app.Use(logger.New())

	// 静态文件服务
	os.MkdirAll("./data/uploads", 0755)
	app.Static("/uploads", "./data/uploads")

	// 支持Range请求的媒体文件服务
	app.Get("/media/*", handlers.ServeMediaFile)

	// Handlers
	authHandler := handlers.NewAuthHandler(db)
	channelHandler := handlers.NewChannelHandler(db, wsHub)
	noteHandler := handlers.NewNoteHandler(db, wsHub)
	fileHandler := handlers.NewFileHandler(db)
	aiHandler := handlers.NewAIHandler(db)

	// WebSocket 路由
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// 获取用户ID
		userIDStr := c.Query("userId")
		if userIDStr == "" {
			c.Close()
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.Close()
			return
		}

		// 生成唯一的连接ID
		connID := fmt.Sprintf("%s-%d", userIDStr, time.Now().UnixNano())

		client := &ws.Client{
			ID:     connID,
			Conn:   c,
			Send:   make(chan []byte, 256),
			UserID: uint(userID),
		}

		wsHub.Register(client)
		defer wsHub.Unregister(client)

		// 启动读写管道
		go client.WritePump()
		client.ReadPump(wsHub)
	}))

	// 协同编辑 WebSocket 路由
	app.Get("/ws/collab", websocket.New(func(c *websocket.Conn) {
		collab.HandleCollabWebSocket(c, yjsServer)
	}))

	// 路由组
	r := app.Group("/api")

	// 公共路由 (无需登录)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	r.Post("/auth/change-password", authHandler.ChangePassword)
	r.Get("/public/notes", noteHandler.GetPublicNotes)

	// 可选认证路由 (可选登录，支持访客和登录用户访问)
	optional := r.Group("/", middleware.OptionalAuth)
	optional.Get("/public/channels", channelHandler.GetPublicChannels)
	optional.Get("/channels/:id", channelHandler.GetChannel)
	optional.Get("/channels/:id/messages", channelHandler.GetChannelMessages)
	optional.Get("/notes/search", noteHandler.SearchNotes) // 搜索路由必须在notes/:id之前
	optional.Get("/notes/:id", noteHandler.GetNote)
	optional.Get("/notes", noteHandler.GetNotes) // 允许访客查看公开笔记

	// 私有路由 (需要登录)
	protected := r.Group("/", middleware.AuthRequired)

	protected.Get("/me", authHandler.GetMe)
	protected.Put("/me", authHandler.UpdateMe)

	protected.Post("/channels", channelHandler.CreateChannel)
	protected.Get("/channels", channelHandler.GetUserChannels)
	protected.Post("/channels/:id/messages", channelHandler.CreateChannelMessage)
	protected.Put("/channels/:id", channelHandler.UpdateChannel)
	protected.Delete("/channels/:id", channelHandler.DeleteChannel)
	protected.Delete("/channels/:id/messages/:messageId", channelHandler.DeleteChannelMessage)
	protected.Put("/channels/:id/messages/:messageId/highlight", channelHandler.HighlightMessage)
	protected.Post("/channels/invite", channelHandler.InviteUser)
	protected.Put("/channels/:id/members/:userId", channelHandler.UpdateMemberRole)
	protected.Delete("/channels/:id/members/:userId", channelHandler.RemoveMember)
	protected.Post("/channels/:id/join", channelHandler.JoinChannelRequest)
	protected.Post("/channels/approvals", channelHandler.HandleMemberStatus)
	protected.Post("/channels/approvals/approve", channelHandler.HandleMemberStatus)
	protected.Get("/channels/approvals/pending", channelHandler.GetPendingApprovals)
	protected.Post("/channels/approvals/:id/accept", channelHandler.AcceptInvitation)
	protected.Delete("/channels/approvals/:id", channelHandler.DeleteApproval)

	protected.Get("/notes", noteHandler.GetNotes)
	protected.Post("/notes", noteHandler.CreateNote)
	protected.Put("/notes/:id", noteHandler.UpdateNote)
	protected.Delete("/notes/:id", noteHandler.DeleteNote)

	protected.Post("/upload", fileHandler.Upload)

	// AI 配置管理路由（仅管理员）
	admin := r.Group("/admin", middleware.AuthRequired, middleware.AdminRequired)
	admin.Get("/ai-config", aiHandler.GetAIConfig)
	admin.Put("/ai-config", aiHandler.UpdateAIConfig)
	admin.Get("/stats", authHandler.GetStats)
	admin.Get("/users", authHandler.GetAllUsers)
	admin.Put("/users/:id/role", authHandler.UpdateUserRole)
	admin.Delete("/users/:id", authHandler.DeleteUser)
	admin.Get("/notes", noteHandler.GetAllNotes)
	admin.Delete("/notes/:id", noteHandler.AdminDeleteNote)
	admin.Get("/channels", handlers.GetAllChannels)
	admin.Put("/channels/:id/public", handlers.AdminToggleChannelPublic)

	// AI 总结路由（已登录用户）
	protected.Post("/ai/summarize", aiHandler.SummarizeNote)
	protected.Post("/ai/polish", aiHandler.PolishNote)

	log.Fatal(app.Listen(":3000"))
}
