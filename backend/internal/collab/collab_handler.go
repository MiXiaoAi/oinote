package collab

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/websocket/v2"
)

// CollabMessage 表示协同编辑消息
type CollabMessage struct {
	Type      string          `json:"type"`      // "sync-step1", "sync-step2", "update", "awareness"
	NoteID    uint            `json:"noteId"`
	Data      json.RawMessage `json:"data"`
	ClientID  string          `json:"clientId,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

// CollabClient 表示协同编辑客户端
type CollabClient struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uint
	Username string
	NoteID   uint
	Server   *YjsServer
}

// NewCollabClient 创建新的协同编辑客户端
func NewCollabClient(conn *websocket.Conn, userID uint, username string, noteID uint, server *YjsServer) *CollabClient {
	clientID := fmt.Sprintf("%d-%d-%d", userID, noteID, time.Now().UnixNano())
	
	return &CollabClient{
		ID:       clientID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
		NoteID:   noteID,
		Server:   server,
	}
}

// ReadPump 从 WebSocket 读取消息
func (c *CollabClient) ReadPump() {
	defer func() {
		// 广播用户离开消息
		c.broadcastUserLeft()
		c.Server.RemoveClient(c.NoteID, c.ID)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump 向 WebSocket 写入消息
func (c *CollabClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 直接发送消息，不批量处理
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// handleMessage 处理接收到的消息
func (c *CollabClient) handleMessage(message []byte) {
	var msg CollabMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("解析协同编辑消息失败: %v", err)
		return
	}

	switch msg.Type {
	case "sync-step1":
		// 客户端请求同步 - 发送状态向量
		c.handleSyncStep1(msg)
		
	case "sync-step2":
		// 客户端发送状态向量，请求差异更新
		c.handleSyncStep2(msg)
		
	case "update":
		// 客户端发送更新
		c.handleUpdate(msg)
		
	case "awareness":
		// 客户端发送感知状态（光标位置等）
		c.handleAwareness(msg)
		
	default:
		log.Printf("未知的消息类型: %s", msg.Type)
	}
}

// handleSyncStep1 处理同步步骤1
func (c *CollabClient) handleSyncStep1(msg CollabMessage) {
	// 发送当前文档的状态向量给客户端
	stateVector := c.Server.GetStateVector(c.NoteID)
	
	response := map[string]interface{}{
		"type":        "sync-step1",
		"stateVector": stateVector,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("序列化 sync-step1 响应失败: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		log.Printf("发送 sync-step1 响应失败")
	}
}

// handleSyncStep2 处理同步步骤2
func (c *CollabClient) handleSyncStep2(msg CollabMessage) {
	// 解析客户端的状态向量
	var data struct {
		StateVector []byte `json:"stateVector"`
	}
	
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Printf("解析状态向量失败: %v", err)
		return
	}

	// 生成差异更新
	update := c.Server.HandleSyncStep1(c.NoteID, c.ID, data.StateVector)
	
	if update != nil {
		response := map[string]interface{}{
			"type":   "sync-step2",
			"update": update,
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			log.Printf("序列化 sync-step2 响应失败: %v", err)
			return
		}

		select {
		case c.Send <- responseData:
		default:
			log.Printf("发送 sync-step2 响应失败")
		}
	}
}

// handleUpdate 处理更新
func (c *CollabClient) handleUpdate(msg CollabMessage) {
	var data struct {
		Update []byte `json:"update"` // Go 会自动将 Base64 字符串解码为 []byte
	}
	
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Printf("解析更新数据失败: %v", err)
		return
	}

	// 应用更新并广播
	c.Server.HandleUpdate(c.NoteID, c.ID, data.Update)
}

// handleAwareness 处理感知状态
func (c *CollabClient) handleAwareness(msg CollabMessage) {
	// 解析光标数据
	var data struct {
		Cursor *struct {
			From int `json:"from"`
			To   int `json:"to"`
		} `json:"cursor,omitempty"`
	}
	
	if err := json.Unmarshal(msg.Data, &data); err != nil {
		log.Printf("解析感知数据失败: %v", err)
		return
	}

	// 广播感知状态到其他客户端
	c.broadcastAwareness(data.Cursor)
}

// broadcastAwareness 广播感知状态
func (c *CollabClient) broadcastAwareness(cursor interface{}) {
	doc := c.Server.GetOrCreateDocument(c.NoteID, "")
	
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	message := map[string]interface{}{
		"type":      "awareness",
		"clientId":  c.ID,
		"userId":    c.UserID,
		"username":  c.Username,
		"cursor":    cursor,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化感知消息失败: %v", err)
		return
	}

	for clientID, client := range doc.Clients {
		if clientID != c.ID {
			select {
			case client.Send <- data:
			default:
				log.Printf("发送感知状态到客户端失败: %s", clientID)
			}
		}
	}
}

// broadcastUserLeft 广播用户离开消息
func (c *CollabClient) broadcastUserLeft() {
	doc := c.Server.GetOrCreateDocument(c.NoteID, "")
	
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	message := map[string]interface{}{
		"type":     "user-left",
		"clientId": c.ID,
		"userId":   c.UserID,
		"username": c.Username,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化用户离开消息失败: %v", err)
		return
	}

	for clientID, client := range doc.Clients {
		if clientID != c.ID {
			select {
			case client.Send <- data:
			default:
				log.Printf("发送用户离开消息到客户端失败: %s", clientID)
			}
		}
	}
}

// HandleCollabWebSocket 处理协同编辑 WebSocket 连接
func HandleCollabWebSocket(conn *websocket.Conn, server *YjsServer) {
	// 从查询参数获取用户信息和笔记ID
	userIDStr := conn.Query("userId")
	username := conn.Query("username")
	noteIDStr := conn.Query("noteId")

	if userIDStr == "" || noteIDStr == "" {
		log.Printf("缺少必要参数: userId=%s, noteId=%s", userIDStr, noteIDStr)
		conn.Close()
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Printf("无效的用户ID: %s", userIDStr)
		conn.Close()
		return
	}

	noteID, err := strconv.ParseUint(noteIDStr, 10, 64)
	if err != nil {
		log.Printf("无效的笔记ID: %s", noteIDStr)
		conn.Close()
		return
	}

	// 创建客户端
	client := NewCollabClient(conn, uint(userID), username, uint(noteID), server)
	
	// 注册客户端到服务器
	server.AddClient(uint(noteID), client.ID, uint(userID), username, client.Send)
	
	// 发送欢迎消息和活跃客户端列表
	activeClients := server.GetActiveClients(uint(noteID))
	welcomeMsg := map[string]interface{}{
		"type":          "welcome",
		"clientId":      client.ID,
		"activeClients": activeClients,
	}
	
	welcomeData, _ := json.Marshal(welcomeMsg)
	client.Send <- welcomeData

	// 启动读写协程
	go client.WritePump()
	client.ReadPump()
}
