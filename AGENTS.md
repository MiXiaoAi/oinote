# iFlow CLI 项目上下文文件

## 项目概述

**oinote** 是一个现代化的协作笔记平台，支持多用户实时协作、富文本编辑、文件共享等功能。

### 核心技术栈

**前端:**
- Vue 3.5.24 (Composition API)
- Pinia 3.0.4 (状态管理)
- Vue Router 4.6.4
- TipTap 3.18.0 (富文本编辑器)
- DaisyUI 5.5.14 + Tailwind CSS 4.1.18
- Vite 7.2.4 (构建工具)
- Axios 1.13.4

**后端:**
- Go 1.25.6
- Fiber 2.52.10 (Web 框架)
- GORM 1.31.1 (ORM)
- SQLite (数据库)
- JWT 5.3.1 (认证)
- WebSocket (实时通信)

### 项目架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue 3 前端     │────│  Go Fiber 后端  │────│  SQLite 数据库  │
│  :5173          │    │  :3000          │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 快速启动

### 环境要求
- Node.js 18+
- Go 1.25+
- Git

### 启动后端
```bash
cd backend
go mod download
go run main.go
```
后端服务将在 `http://localhost:3000` 启动

### 启动前端
```bash
cd frontend
pnpm install  # 或 npm install
pnpm dev      # 或 npm run dev
```
前端服务将在 `http://localhost:5173` 启动

### 构建前端
```bash
cd frontend
pnpm build  # 或 npm run build
```

## 项目结构

```
oinote/
├── backend/                    # Go 后端
│   ├── api/                   # API 处理器
│   │   ├── auth.go           # 认证相关 API
│   │   ├── channel.go        # 频道相关 API
│   │   ├── channels.go       # 频道列表 API
│   │   ├── files.go          # 文件服务 API
│   │   ├── media.go          # 媒体文件服务
│   │   ├── note.go           # 单个笔记 API
│   │   ├── notes.go          # 笔记列表 API
│   │   ├── upload.go         # 文件上传 API
│   │   └── ai.go             # AI 功能 API
│   ├── config/               # 配置
│   │   └── db.go             # 数据库配置
│   ├── data/                 # 数据目录
│   │   ├── oinote.db         # SQLite 数据库
│   │   └── uploads/          # 上传文件目录
│   │       ├── avatars/      # 头像
│   │       └── channels/     # 频道文件
│   ├── internal/             # 内部模块
│   │   ├── middleware/       # 中间件
│   │   │   └── auth.go       # JWT 认证中间件
│   │   ├── models/           # 数据模型
│   │   │   └── schema.go     # 数据库模型定义
│   │   ├── utils/            # 工具函数
│   │   │   └── jwt.go        # JWT 工具
│   │   └── websocket/        # WebSocket
│   │       └── hub.go        # WebSocket Hub
│   ├── main.go               # 应用入口
│   └── go.mod                # Go 模块配置
│
├── frontend/                  # Vue 3 前端
│   ├── src/
│   │   ├── api/              # API 调用
│   │   │   └── axios.js      # Axios 实例配置
│   │   ├── components/       # Vue 组件
│   │   │   ├── ManageModal.vue    # 管理模态框
│   │   │   ├── MediaPlayer.vue    # 媒体播放器
│   │   │   ├── MobileDrawer.vue   # 移动端抽屉
│   │   │   └── Sidebar.vue        # 侧边栏
│   │   ├── data/             # 数据文件
│   │   │   └── quotes.txt    # 引用语
│   │   ├── layouts/          # 布局组件
│   │   │   └── MainLayout.vue
│   │   ├── router/           # 路由配置
│   │   │   └── index.js
│   │   ├── stores/           # Pinia 状态管理
│   │   │   ├── auth.js       # 认证状态
│   │   │   └── theme.js      # 主题状态
│   │   ├── utils/            # 工具函数
│   │   │   ├── eventBus.js   # 事件总线
│   │   │   ├── urlHelper.js  # URL 辅助
│   │   │   └── websocket.js  # WebSocket 客户端
│   │   ├── views/            # 页面视图
│   │   │   ├── Auth/
│   │   │   │   ├── Login.vue
│   │   │   │   └── Register.vue
│   │   │   ├── Home.vue      # 首页
│   │   │   ├── NotesList.vue
│   │   │   ├── NoteEditor.vue
│   │   │   ├── ChannelsList.vue
│   │   │   ├── ChannelView.vue
│   │   │   └── Approvals.vue # 审批页面
│   │   ├── App.vue           # 根组件
│   │   ├── main.js           # 应用入口
│   │   └── style.css         # 全局样式
│   ├── index.html            # HTML 入口
│   ├── package.json          # 依赖配置
│   └── vite.config.js        # Vite 配置
│
├── .gitignore               # Git 忽略配置
└── README.md                # 项目说明
```

## 数据库模型

### User（用户）
- `id` - 用户 ID
- `username` - 用户名（唯一）
- `password` - 密码（bcrypt 哈希）
- `nickname` - 昵称
- `avatar` - 头像 URL
- `bio` - 个人简介
- `role` - 角色（'admin', 'member'）

### Channel（频道）
- `id` - 频道 ID
- `name` - 频道名称
- `description` - 频道描述
- `owner_id` - 所有者 ID
- `is_public` - 是否公开
- `theme_color` - 主题颜色（默认 '#87CEEB'）
- `tags` - 标签（逗号分隔）

### ChannelMember（频道成员）
- `id` - 成员 ID
- `channel_id` - 频道 ID
- `user_id` - 用户 ID
- `role` - 角色（'owner', 'admin', 'member'）
- `status` - 状态（'active', 'pending', 'invited'）
- `joined_at` - 加入时间

### Note（笔记）
- `id` - 笔记 ID
- `title` - 标题
- `content` - 内容（HTML）
- `channel_id` - 频道 ID（null 为个人笔记）
- `owner_id` - 作者 ID
- `is_public` - 是否公开
- `tags` - 标签（逗号分隔）
- `line_spacing` - 行间距（默认 1.5）

### Attachment（附件）
- `id` - 附件 ID
- `file_name` - 文件名
- `file_path` - 文件路径
- `file_size` - 文件大小
- `file_type` - 文件类型
- `uploader_id` - 上传者 ID
- `channel_id` - 频道 ID（可选）
- `note_id` - 笔记 ID（可选）

### ChannelMessage（频道消息）
- `id` - 消息 ID
- `channel_id` - 频道 ID
- `user_id` - 用户 ID
- `content` - 消息内容
- `attachment_id` - 附件 ID（可选）
- `is_highlighted` - 是否高亮

### AIConfig（AI 配置）
- `id` - 配置 ID
- `openai_url` - OpenAI API URL
- `api_key` - API Key
- `model` - 模型名称
- `updated_by` - 最后更新的管理员 ID

## 主要 API 路由

### 公共路由（无需认证）
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录
- `POST /api/auth/change-password` - 修改密码
- `GET /api/public/notes` - 获取公开笔记

### 可选认证路由（访客或登录用户）
- `GET /api/public/channels` - 获取公开频道
- `GET /api/channels/:id` - 获取频道详情
- `GET /api/channels/:id/messages` - 获取频道消息
- `GET /api/notes/search` - 搜索笔记和频道
- `GET /api/notes/:id` - 获取笔记详情
- `GET /api/notes` - 获取笔记列表

### 私有路由（需要登录）
- `GET /api/me` - 获取当前用户信息
- `PUT /api/me` - 更新当前用户信息
- `GET /api/channels` - 获取用户的频道列表
- `POST /api/channels` - 创建频道
- `PUT /api/channels/:id` - 更新频道
- `DELETE /api/channels/:id` - 删除频道
- `POST /api/channels/:id/messages` - 发送消息
- `DELETE /api/channels/:id/messages/:messageId` - 删除消息
- `PUT /api/channels/:id/messages/:messageId/highlight` - 高亮消息
- `POST /api/channels/invite` - 邀请用户
- `PUT /api/channels/:id/members/:userId` - 更新成员角色
- `DELETE /api/channels/:id/members/:userId` - 移除成员
- `POST /api/channels/:id/join` - 申请加入频道
- `POST /api/channels/approvals` - 处理成员申请
- `GET /api/channels/approvals/pending` - 获取待审批列表
- `POST /api/channels/approvals/:id/accept` - 接受邀请
- `DELETE /api/channels/approvals/:id` - 拒绝申请
- `POST /api/notes` - 创建笔记
- `PUT /api/notes/:id` - 更新笔记
- `DELETE /api/notes/:id` - 删除笔记
- `POST /api/upload` - 上传文件
- `GET /media/*` - 获取媒体文件（支持 Range 请求）
- `/uploads/*` - 静态文件服务
- `GET /ws` - WebSocket 连接（需传递 userId 参数）
- `POST /api/ai/summarize` - AI 总结笔记
- `POST /api/ai/polish` - AI 润色笔记

### 管理员路由
- `GET /api/admin/ai-config` - 获取 AI 配置
- `PUT /api/admin/ai-config` - 更新 AI 配置
- `GET /api/admin/stats` - 获取系统统计
- `GET /api/admin/users` - 获取所有用户
- `PUT /api/admin/users/:id/role` - 更新用户角色
- `DELETE /api/admin/users/:id` - 删除用户
- `GET /api/admin/notes` - 获取所有笔记
- `DELETE /api/admin/notes/:id` - 管理员删除笔记
- `GET /api/admin/channels` - 获取所有频道
- `PUT /api/admin/channels/:id/public` - 切换频道公开状态

## 前端路由

- `/` - 首页（公开频道和笔记）
- `/notes` - 笔记列表
- `/channels` - 频道列表
- `/channel/:id` - 频道详情（聊天/笔记模式切换）
- `/note/:id?` - 笔记编辑器
- `/approvals` - 审批页面（需要登录）
- `/login` - 登录页
- `/register` - 注册页

## 开发规范

### 前端
- 使用 Vue 3 Composition API
- 使用 Pinia 进行状态管理
- 使用 Tailwind CSS 和 DaisyUI
- 使用 Axios 进行 API 调用
- TipTap 编辑器默认左对齐，图片默认宽度 20%
- 使用 `@` 别名引用 `src` 目录

### 后端
- 使用 GORM 进行数据库操作
- 使用 Fiber 的 Context 处理请求
- 统一错误处理和响应格式
- 使用 JWT 进行认证
- BodyLimit 配置为 2GB
- 笔记附件限制为 100MB

### Git
- 分支命名：`feature/xxx`, `bugfix/xxx`
- 提交信息使用中文：`feat: 新功能`, `fix: 修复bug`
- 忽略文件：依赖、构建输出、数据库、上传文件、备份文件

## 重要配置

### 前端配置
- 开发端口：5173
- API 地址：`http://localhost:3000/api`
- 使用 Vite 构建
- 支持热模块替换（HMR）

### 后端配置
- 端口：3000
- 数据库：`backend/data/oinote.db`
- 上传目录：`backend/data/uploads/`
- CORS 允许所有来源

## 用户偏好

- 用户偏好使用 **pnpm** 包管理器（而不是 npm）

## 安全特性

- JWT 认证
- 密码 bcrypt 哈希加密
- 输入验证（前后端双重验证）
- 文件类型检查和路径验证
- 基于角色的访问控制
- CORS 配置

## 关键文件说明

### 后端核心文件
- `backend/main.go` - 应用入口，路由配置
- `backend/internal/models/schema.go` - 数据库模型定义
- `backend/internal/middleware/auth.go` - JWT 认证中间件
- `backend/config/db.go` - 数据库连接配置
- `backend/internal/websocket/hub.go` - WebSocket Hub

### 前端核心文件
- `frontend/src/main.js` - 应用入口
- `frontend/src/App.vue` - 根组件
- `frontend/src/router/index.js` - 路由配置
- `frontend/src/api/axios.js` - Axios 实例
- `frontend/src/stores/auth.js` - 认证状态管理
- `frontend/vite.config.js` - Vite 配置

## 测试

前端使用 Playwright 进行测试：
```bash
cd frontend
pnpm test  # 或 npm test
```

## 项目链接

- **GitHub**: https://github.com/MiXiaoAi/oinote
- **作者**: MiXiaoAi