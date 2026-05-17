package collab

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	y "github.com/skyterra/y-crdt"
)

// YjsDocument 表示一个协同编辑文档
type YjsDocument struct {
	Doc         *y.Doc
	LastUpdated time.Time
	Clients     map[string]*YjsClient
	mu          sync.RWMutex
}

// YjsClient 表示连接到文档的客户端
type YjsClient struct {
	ClientID string
	UserID   uint
	Username string
	Send     chan []byte
}

// YjsServer 管理所有协同编辑文档
type YjsServer struct {
	documents map[uint]*YjsDocument // noteID -> YjsDocument
	mu        sync.RWMutex
}

// SyncMessage 表示同步消息
type SyncMessage struct {
	Type      string          `json:"type"`      // "sync", "awareness", "update"
	NoteID    uint            `json:"noteId"`
	Data      json.RawMessage `json:"data"`
	ClientID  string          `json:"clientId"`
	Timestamp int64           `json:"timestamp"`
}

// AwarenessState 表示用户状态
type AwarenessState struct {
	User struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Color    string `json:"color"`
	} `json:"user"`
	Cursor *struct {
		Anchor int `json:"anchor"`
		Head   int `json:"head"`
	} `json:"cursor,omitempty"`
}

// NewYjsServer 创建新的 Yjs 服务器
func NewYjsServer() *YjsServer {
	server := &YjsServer{
		documents: make(map[uint]*YjsDocument),
	}
	
	// 启动清理任务
	go server.cleanupInactiveDocuments()
	
	return server
}

// GetOrCreateDocument 获取或创建文档
func (s *YjsServer) GetOrCreateDocument(noteID uint, initialContent string) *YjsDocument {
	s.mu.Lock()
	defer s.mu.Unlock()

	if doc, exists := s.documents[noteID]; exists {
		doc.LastUpdated = time.Now()
		return doc
	}

	// 创建新文档
	ydoc := y.NewDoc(fmt.Sprintf("%d", noteID), false, nil, nil, false)
	
	// 如果有初始内容，设置到文档中
	if initialContent != "" {
		ydoc.Transact(func(trans *y.Transaction) {
			ytext := ydoc.GetText("content")
			ytext.Insert(0, initialContent, nil)
		}, nil)
	}

	doc := &YjsDocument{
		Doc:         ydoc,
		LastUpdated: time.Now(),
		Clients:     make(map[string]*YjsClient),
	}

	s.documents[noteID] = doc
	
	return doc
}

// AddClient 添加客户端到文档
func (s *YjsServer) AddClient(noteID uint, clientID string, userID uint, username string, send chan []byte) {
	doc := s.GetOrCreateDocument(noteID, "")
	
	doc.mu.Lock()
	defer doc.mu.Unlock()

	client := &YjsClient{
		ClientID: clientID,
		UserID:   userID,
		Username: username,
		Send:     send,
	}

	doc.Clients[clientID] = client
}

// RemoveClient 从文档移除客户端
func (s *YjsServer) RemoveClient(noteID uint, clientID string) {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	doc.mu.Lock()
	defer doc.mu.Unlock()

	delete(doc.Clients, clientID)
}

// HandleSyncStep1 处理同步步骤1 - 客户端请求状态向量
func (s *YjsServer) HandleSyncStep1(noteID uint, clientID string, stateVector []byte) []byte {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	// 编码当前文档状态作为更新
	var update []byte
	if len(stateVector) > 0 {
		// 根据客户端的状态向量生成差异更新
		targetStateVector := y.DecodeStateVector(stateVector)
		update = y.EncodeStateAsUpdate(doc.Doc, y.EncodeStateVector(doc.Doc, targetStateVector, y.NewUpdateEncoderV1()))
	} else {
		// 发送完整文档状态
		update = y.EncodeStateAsUpdate(doc.Doc, nil)
	}

	return update
}

// HandleUpdate 处理客户端更新
func (s *YjsServer) HandleUpdate(noteID uint, clientID string, update []byte) {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	// 直接应用更新
	y.ApplyUpdate(doc.Doc, update, nil)
	doc.LastUpdated = time.Now()

	// 广播更新到其他客户端
	s.broadcastUpdate(doc, clientID, update)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// broadcastUpdate 广播更新到所有其他客户端
func (s *YjsServer) broadcastUpdate(doc *YjsDocument, senderClientID string, update []byte) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	message := map[string]interface{}{
		"type":   "update",
		"update": update,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化更新消息失败: %v", err)
		return
	}

	for clientID, client := range doc.Clients {
		if clientID != senderClientID {
			select {
			case client.Send <- data:
			default:
				log.Printf("发送更新到客户端失败: %s", clientID)
			}
		}
	}
}

// GetDocumentContent 获取文档内容
func (s *YjsServer) GetDocumentContent(noteID uint) string {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return ""
	}

	ytext := doc.Doc.GetText("content")
	return ytext.ToString()
}

// GetStateVector 获取文档状态向量
func (s *YjsServer) GetStateVector(noteID uint) []byte {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	stateVector := y.GetStateVector(doc.Doc.Store)
	return y.EncodeStateVector(doc.Doc, stateVector, y.NewUpdateEncoderV1())
}

// cleanupInactiveDocuments 清理不活跃的文档
func (s *YjsServer) cleanupInactiveDocuments() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		
		for noteID, doc := range s.documents {
			doc.mu.RLock()
			clientCount := len(doc.Clients)
			lastUpdated := doc.LastUpdated
			doc.mu.RUnlock()

			// 如果文档没有客户端且超过30分钟未更新，则清理
			if clientCount == 0 && now.Sub(lastUpdated) > 30*time.Minute {
				delete(s.documents, noteID)
				log.Printf("清理不活跃文档: noteID=%d", noteID)
			}
		}
		
		s.mu.Unlock()
	}
}

// GetActiveClients 获取文档的活跃客户端列表
func (s *YjsServer) GetActiveClients(noteID uint) []map[string]interface{} {
	s.mu.RLock()
	doc, exists := s.documents[noteID]
	s.mu.RUnlock()

	if !exists {
		return []map[string]interface{}{}
	}

	doc.mu.RLock()
	defer doc.mu.RUnlock()

	clients := make([]map[string]interface{}, 0, len(doc.Clients))
	for _, client := range doc.Clients {
		clients = append(clients, map[string]interface{}{
			"clientId": client.ClientID,
			"userId":   client.UserID,
			"username": client.Username,
		})
	}

	return clients
}
