package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Client 表示一个 WebSocket 客户端
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	UserID uint
}

// Message 表示要广播的消息
type Message struct {
	Type      string      `json:"type"`      // "note", "channel", "channel_message"
	Action    string      `json:"action"`    // "create", "update", "delete"
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Hub 管理 WebSocket 连接
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 从客户端接收的消息
	broadcast chan []byte

	// 注册客户端请求
	register chan *Client

	// 注销客户端请求
	unregister chan *Client
}

// NewHub 创建一个新的 Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run 启动 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket 客户端已连接: %s (用户ID: %d)", client.ID, client.UserID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("WebSocket 客户端已断开: %s", client.ID)
			}

		case message := <-h.broadcast:
			// 向所有连接的客户端广播消息
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// 发送失败，关闭连接
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastMessage 广播消息到所有客户端
func (h *Hub) BroadcastMessage(msgType, action string, data interface{}) {
	message := Message{
		Type:      msgType,
		Action:    action,
		Data:      data,
		Timestamp: 0,
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket 广播消息失败: %v", err)
		return
	}

	h.broadcast <- bytes
}

// WritePump 从 Hub 读取消息并写入 WebSocket 连接
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Hub 关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 排队消息
			for len(c.Send) > 0 {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 从 WebSocket 连接读取消息
func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			break
		}
	}
}
