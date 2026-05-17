package models

import (
	"time"
)

// 成员状态常量
const (
	MemberStatusActive  = "active"  // 正式成员
	MemberStatusInvited = "invited" // 被邀请，等待用户同意
	MemberStatusPending = "pending" // 申请加入，等待管理员批准
)

// 角色常量
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // 密码不输出
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Role     string `gorm:"default:'member'" json:"role"` // admin, member
}

type Channel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	OwnerID     uint   `gorm:"not null" json:"owner_id"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
	ThemeColor  string `gorm:"default:'#87CEEB'" json:"theme_color"` // 天蓝色
	Tags        string `json:"tags"`                                 // 逗号分隔
	MemberCount int    `gorm:"-" json:"member_count"`                // 成员数（不从数据库加载）

	// 关联
	Owner   User            `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []ChannelMember `json:"members,omitempty"`
}

type ChannelMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID uint      `gorm:"index" json:"channel_id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Role      string    `gorm:"default:'member'" json:"role"`   // owner, admin, member
	Status    string    `gorm:"default:'active'" json:"status"` // active, invited, pending
	JoinedAt  time.Time `json:"joined_at"`

	// 预加载用户信息
	User User `gorm:"foreignKey:UserID" json:"user"`
	// 预加载频道信息
	Channel Channel `gorm:"foreignKey:ChannelID" json:"channel"`
}

type Note struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Title        string `json:"title"`
	Content      string `gorm:"type:text" json:"content"`
	ChannelID    *uint  `gorm:"index" json:"channel_id"`   // null 为个人笔记
	OwnerID      uint   `gorm:"index" json:"owner_id"`
	IsPublic     bool   `gorm:"default:false" json:"is_public"`
	Tags         string `json:"tags"`           // 逗号分隔
	LineSpacing  float64 `gorm:"default:1.5" json:"line_spacing"` // 行间距

	// 关联
	Owner User `gorm:"foreignKey:OwnerID" json:"owner"`
}

type Attachment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	FileSize   int64  `json:"file_size"`
	FileType   string `json:"file_type"`
	UploaderID uint   `json:"uploader_id"`
	ChannelID  *uint  `json:"channel_id"`
	NoteID     *uint  `json:"note_id"`
}

type ChannelMessage struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ChannelID     uint      `gorm:"index" json:"channel_id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	Content       string    `gorm:"type:text" json:"content"`
	AttachmentID  *uint     `json:"attachment_id"`
	IsHighlighted bool      `gorm:"default:false" json:"is_highlighted"`

	User       User       `gorm:"foreignKey:UserID" json:"user"`
	Attachment Attachment `gorm:"foreignKey:AttachmentID" json:"attachment"`
}

type AIConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OpenAIURL  string `json:"openai_url"`  // OpenAI API URL
	APIKey     string `json:"api_key"`     // API Key
	Model      string `json:"model"`       // Model name (e.g., gpt-4, gpt-3.5-turbo)
	UpdatedBy  uint   `json:"updated_by"`  // 最后更新的管理员ID
}
