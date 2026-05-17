<div align="center">
  <h1>oinote</h1>
  <p>类 Notion 的协作文档平台 · 实时编辑 · AI 辅助</p>

  <p>
    <img src="https://img.shields.io/badge/Frontend-Vue%203%20%7C%20Vite%20%7C%20Tailwind%20v4-4FC08D" />
    <img src="https://img.shields.io/badge/Backend-Go%20%7C%20Fiber%20%7C%20GORM-00ADD8" />
    <img src="https://img.shields.io/badge/Database-SQLite-003B57" />
    <img src="https://img.shields.io/badge/Collab-Yjs%20%7C%20WebSocket-FF6B6B" />
  </p>
</div>

---

## 功能

- **多频道** — 创建频道、管理成员、频道审批
- **富文本编辑** — Tiptap 编辑器，支持表格、图片、链接、排版
- **实时协同** — Yjs + WebSocket 多用户同时编辑，远程光标
- **AI 辅助** — AI 摘要、续写、翻译、润色（支持 OpenAI 兼容接口）
- **媒体管理** — 图片/视频/音频上传与播放
- **审批系统** — 用户申请加入频道的审批流
- **深色模式** — DaisyUI 主题切换
- **移动端适配** — 响应式布局 + 侧边抽屉

## 快速开始

### 前置要求

- Go 1.25+
- Node.js 20+

### 启动

```bash
# 1. 启动后端 (http://localhost:3000)
cd backend
go run main.go

# 2. 另开终端，启动前端 (http://localhost:5173)
cd frontend
npm install
npm run dev
```

首次运行会自动创建 SQLite 数据库并生成默认管理员账号：

```
用户名: admin
密码:   admin
```

## 项目结构

```
oinote/
├── backend/                  # Go 后端
│   ├── main.go               # 入口 + 路由注册
│   ├── api/                  # HTTP 处理器
│   │   ├── auth.go           # 登录/注册/用户管理
│   │   ├── note.go           # 笔记 CRUD
│   │   ├── notes.go          # 笔记列表/搜索
│   │   ├── channel.go        # 单频道操作
│   │   ├── channels.go       # 频道列表/管理
│   │   ├── ai.go             # AI 对话
│   │   ├── files.go          # 文件服务
│   │   ├── media.go          # 媒体元数据
│   │   └── upload.go         # 文件上传
│   ├── config/
│   │   └── db.go             # SQLite 连接 + 自动迁移
│   └── internal/
│       ├── middleware/
│       │   └── auth.go       # JWT 验证中间件
│       ├── models/
│       │   └── schema.go     # GORM 数据模型
│       ├── collab/           # Yjs 协同编辑
│       └── websocket/        # WebSocket Hub
│
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── views/            # 页面组件
│   │   ├── components/       # 通用组件
│   │   ├── stores/           # Pinia 状态
│   │   ├── router/           # 路由
│   │   ├── api/              # Axios 实例
│   │   └── utils/            # 工具函数
│   └── vite.config.js
│
└── README.md
```

## API 概览

| 路径 | 说明 |
|------|------|
| `POST /api/auth/login` | 用户登录 |
| `POST /api/auth/register` | 用户注册 |
| `GET/PUT /api/auth/users/:id` | 用户信息 |
| `GET/POST /api/channels` | 频道列表 / 创建 |
| `GET/PUT/DELETE /api/channels/:id` | 频道详情 |
| `POST /api/channels/:id/join` | 申请加入 |
| `GET/POST /api/channels/:id/notes` | 笔记列表 / 创建 |
| `GET/PUT/DELETE /api/notes/:id` | 笔记 CRUD |
| `GET /api/notes/search` | 搜索笔记 |
| `POST /api/ai/chat` | AI 对话 |
| `POST /api/upload` | 文件上传 |
| `WS /ws` | 通知 WebSocket |
| `WS /ws/collab` | Yjs 协同 WebSocket |

## 技术栈

**前端**
- [Vue 3](https://vuejs.org/) + [Vite](https://vitejs.dev/)
- [Tailwind CSS v4](https://tailwindcss.com/) + [DaisyUI v5](https://daisyui.com/)
- [Tiptap](https://tiptap.dev/) 富文本编辑器
- [Pinia](https://pinia.vuejs.org/) 状态管理
- [Yjs](https://yjs.dev/) CRDT 协同同步

**后端**
- [Go](https://go.dev/) + [Fiber](https://gofiber.io/)
- [GORM](https://gorm.io/) ORM
- [SQLite](https://sqlite.org/) (纯 Go 驱动 [glebarez/sqlite](https://github.com/glebarez/sqlite))
- [JWT](https://github.com/golang-jwt/jwt) 认证
- [WebSocket](https://github.com/gofiber/websocket) 实时通信
