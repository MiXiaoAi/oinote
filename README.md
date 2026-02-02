# oinote 协作笔记平台

> 一个现代化的协作笔记平台，支持多用户实时协作、富文本编辑、文件共享等功能。

## 项目特色

- **现代化技术栈** - Vue 3 + Go Fiber + SQLite
- **富文本编辑** - 基于 TipTap 的强大编辑器
- **实时协作** - 频道聊天、消息管理、权限控制
- **响应式设计** - 完美适配桌面和移动设备
- **媒体支持** - 优化的视频播放和文件处理
- **访客模式** - 公开内容访问，友好的用户体验

## 快速开始

### 环境要求

- Node.js 18+
- Go 1.25+
- Git

### 安装运行

#### 1. 克隆项目
```bash
git clone https://github.com/MiXiaoAi/oinote.git
cd oinote
```

#### 2. 启动后端
```bash
cd backend
go mod download
go run main.go
```
后端服务将在 `http://localhost:3000` 启动

#### 3. 启动前端
```bash
cd frontend
npm install
npm run dev
```
前端服务将在 `http://localhost:5173` 启动

#### 4. 访问应用
打开浏览器访问 `http://localhost:5173`

## 技术架构

### 前端技术栈
- **Vue 3.5.24** - 使用 Composition API 的现代化前端框架
- **Pinia 3.0.4** - 状态管理
- **Vue Router 4.6.4** - 路由管理
- **TipTap 3.18.0** - 富文本编辑器（支持图片、表格、文本对齐、链接等）
- **DaisyUI 5.5.14** - 基于 Tailwind CSS 的 UI 组件库
- **Tailwind CSS 4.1.18** - 实用优先的 CSS 框架
- **Vite 7.2.4** - 快速的构建工具
- **Axios 1.13.4** - HTTP 客户端
- **Lucide Vue Next** - 图标库

### 后端技术栈
- **Go 1.25.6** - 后端编程语言
- **Fiber 2.52.10** - 高性能 Web 框架
- **GORM 1.31.1** - ORM 框架
- **SQLite** - 轻量级数据库
- **JWT 5.3.1** - 无状态令牌认证
- **WebSocket** - 实时通信支持

### 系统架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue 3 前端     │────│  Go Fiber 后端  │────│  SQLite 数据库  │
│                 │    │                 │    │                 │
│ - 用户界面       │    │ - RESTful API   │    │ - 用户数据       │
│ - 状态管理       │    │ - JWT认证        │    │ - 笔记内容       │
│ - 路由管理       │    │ - 文件服务       │    │ - 频道信息       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 项目结构

```
oinote/
├── backend/                 # Go 后端
│   ├── api/                # API 处理器
│   │   ├── auth.go         # 认证相关 API
│   │   ├── channel.go      # 频道相关 API
│   │   ├── channels.go     # 频道列表 API
│   │   ├── files.go        # 文件服务 API
│   │   ├── media.go        # 媒体文件服务
│   │   ├── note.go         # 单个笔记 API
│   │   ├── notes.go        # 笔记列表 API
│   │   └── upload.go       # 文件上传 API
│   ├── config/             # 配置
│   │   └── db.go           # 数据库配置
│   ├── data/               # 数据目录
│   │   ├── oinote.db       # SQLite 数据库
│   │   └── uploads/        # 上传文件目录
│   │       ├── channels/   # 频道文件
│   │       ├── notes/      # 笔记附件
│   │       └── others/     # 其他文件
│   ├── internal/           # 内部模块
│   │   ├── middleware/     # 中间件
│   │   ├── models/         # 数据模型
│   │   ├── utils/          # 工具函数
│   │   └── websocket/      # WebSocket
│   ├── main.go             # 应用入口
│   ├── go.mod              # Go 模块配置
│   └── go.sum              # Go 依赖锁定
│
├── frontend/               # Vue 3 前端
│   ├── src/
│   │   ├── api/            # API 调用
│   │   ├── assets/         # 静态资源
│   │   ├── components/     # Vue 组件
│   │   ├── data/           # 数据文件
│   │   ├── layouts/        # 布局组件
│   │   ├── router/         # 路由配置
│   │   ├── stores/         # Pinia 状态管理
│   │   ├── utils/          # 工具函数
│   │   ├── views/          # 页面视图
│   │   ├── App.vue         # 根组件
│   │   ├── main.js         # 应用入口
│   │   └── style.css       # 全局样式
│   ├── public/             # 公共静态资源
│   ├── index.html          # HTML 入口
│   ├── package.json        # 依赖配置
│   └── vite.config.js      # Vite 配置
│
├── .gitignore              # Git 忽略配置
└── README.md               # 项目说明
```

## 数据库模型

### User（用户）
- `id` - 用户 ID
- `username` - 用户名（唯一）
- `email` - 邮箱（唯一）
- `password` - 密码（bcrypt 哈希）
- `avatar` - 头像 URL
- `created_at` - 创建时间
- `updated_at` - 更新时间

### Channel（频道）
- `id` - 频道 ID
- `name` - 频道名称
- `description` - 频道描述
- `is_public` - 是否公开
- `owner_id` - 所有者 ID
- `created_at` - 创建时间
- `updated_at` - 更新时间

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
- `is_public` - 是否公开
- `author_id` - 作者 ID
- `created_at` - 创建时间
- `updated_at` - 更新时间

### Attachment（附件）
- `id` - 附件 ID
- `file_name` - 文件名
- `file_path` - 文件路径
- `file_size` - 文件大小
- `file_type` - 文件类型（'note_attachment', 'channel_file'）
- `uploader_id` - 上传者 ID
- `channel_id` - 频道 ID（可选）
- `note_id` - 笔记 ID（可选）
- `created_at` - 创建时间
- `updated_at` - 更新时间

### ChannelMessage（频道消息）
- `id` - 消息 ID
- `channel_id` - 频道 ID
- `user_id` - 用户 ID
- `content` - 消息内容
- `attachment_id` - 附件 ID（可选）
- `is_highlighted` - 是否高亮
- `created_at` - 创建时间
- `updated_at` - 更新时间

### API 路由

#### 公共路由（无需认证）
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录（细化错误提示）
- `POST /api/auth/change-password` - 修改密码
- `GET /api/public/notes` - 获取公开笔记

#### 可选认证路由（访客或登录用户）
- `GET /api/notes/search` - 搜索笔记和频道
- `GET /api/public/channels` - 获取公开频道
- `GET /api/channels/:id` - 获取频道详情
- `GET /api/channels/:id/messages` - 获取频道消息
- `GET /api/notes/:id` - 获取笔记详情
- `GET /api/notes` - 获取笔记列表

#### 私有路由（需要登录）
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
- `DELETE /api/channels/:id/members/:userId` - 移除成员
- `POST /api/channels/:id/join` - 申请加入频道
- `POST /api/channels/approvals` - 处理成员申请
- `GET /api/channels/approvals/pending` - 获取待审批列表
- `POST /api/channels/approvals/:id/accept` - 接受邀请
- `DELETE /api/channels/approvals/:id` - 拒绝申请
- `GET /api/notes` - 获取笔记列表
- `POST /api/notes` - 创建笔记
- `PUT /api/notes/:id` - 更新笔记
- `DELETE /api/notes/:id` - 删除笔记
- `POST /api/upload` - 上传文件
- `GET /media/*` - 获取媒体文件（支持 Range 请求）
- `/uploads/*` - 静态文件服务
- `GET /ws` - WebSocket 连接（需传递 userId 参数）

#### 管理员路由
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

#### AI 功能路由
- `POST /api/ai/summarize` - AI 总结笔记
- `POST /api/ai/polish` - AI 润色笔记

## 前端路由

### 主布局路由
- `/` - 首页（公开频道和笔记）
- `/notes` - 笔记列表
- `/channels` - 频道列表
- `/channel/:id` - 频道详情（聊天/笔记模式切换）
- `/note/:id?` - 笔记编辑器
- `/approvals` - 审批页面（需要登录）

### 认证路由
- `/login` - 登录页（含修改密码功能）
- `/register` - 注册页

## 功能介绍

### 笔记管理
- 富文本编辑器，支持图片、表格、链接等
- 笔记分类和标签系统
- 公开/私有笔记设置
- 自动保存和历史记录
- 图片拖拽上传和粘贴上传
- 图片右键菜单编辑和删除
- 图片可调整大小
- 频道笔记管理（右键菜单管理功能）

### 频道协作
- 创建公开或私有频道
- 实时聊天和消息交流
- 文件分享和媒体播放
- 成员管理和权限控制
- 消息高亮功能
- 频道成员申请和审批
- 频道视图模式切换（聊天/笔记）

### 全局搜索
- 实时搜索（输入时自动搜索）
- 搜索范围：笔记和频道
- 搜索字段：标题、内容、标签
- 显示匹配原因（标题匹配、内容匹配、标签匹配）
- 笔记搜索结果显示：标题、作者、是否频道笔记、标签、更新时间
- 频道搜索结果显示：名称、描述、标签

### 用户系统
- 用户注册和登录
- 个人资料管理
- 头像上传和自定义
- 密码修改功能
- 细化的错误提示（用户名不存在/密码错误）
- JWT令牌认证
- 用户角色管理（管理员可更改用户角色）

### 后台管理（仅管理员）
- 系统统计信息
- 用户管理（查看、更改角色、删除）
- 笔记管理（查看、删除）
- 频道管理（查看、切换公开状态、删除）
- AI 配置管理（OpenAI API 设置）

## 开发指南

### 环境配置

#### 后端配置
- 端口：3000
- 数据库：`backend/data/oinote.db`
- 上传目录：`backend/data/uploads/`
- 文件大小限制：2GB（通过 BodyLimit 配置）
- 笔记附件限制：100MB（在代码中限制）

#### 前端配置
- 开发端口：5173
- API 地址：`http://localhost:3000/api`

### 开发流程

1. **功能开发**
   - 后端：在 `handlers/` 添加新的API处理器
   - 前端：在 `src/views/` 添加新的页面组件

2. **数据库迁移**
   - 修改 `internal/models/schema.go` 中的数据模型
   - 启动应用时自动迁移数据库结构

3. **测试验证**
   - 使用浏览器开发者工具调试
   - 查看后端日志确认API调用

### 开发规范

#### 前端
- 使用 Vue 3 Composition API
- 使用 Pinia 进行状态管理
- 使用 Tailwind CSS 和 DaisyUI
- 使用 Axios 进行 API 调用
- TipTap 编辑器默认左对齐，图片默认宽度 20%

#### 后端
- 使用 GORM 进行数据库操作
- 使用 Fiber 的 Context 处理请求
- 统一错误处理和响应格式
- 使用 JWT 进行认证

#### Git
- 分支命名：`feature/xxx`, `bugfix/xxx`
- 提交信息使用中文：`feat: 新功能`, `fix: 修复bug`
- 忽略文件：依赖、构建输出、数据库、上传文件、备份文件

## 安全特性

- **JWT认证** - 无状态令牌认证
- **密码加密** - bcrypt哈希加密
- **输入验证** - 前后端双重验证
- **文件安全** - 类型检查和路径验证
- **权限控制** - 基于角色的访问控制
- **CORS配置** - 允许跨域请求

## 性能指标

| 指标 | 数值 | 说明 |
|------|------|------|
| 页面加载时间 | < 2s | 首屏渲染时间 |
| API响应时间 | < 500ms | 平均接口响应 |
| 并发用户数 | 100+ | 同时在线用户 |
| 文件上传速度 | > 10MB/s | 大文件上传速度 |

## 常见问题

1. **跨域问题**：后端已配置 CORS 允许所有来源
2. **文件上传失败**：检查后端 BodyLimit 配置和目录权限
3. **WebSocket 断连**：检查连接 ID 和用户 ID 传递
4. **数据库迁移失败**：删除 `oinote.db` 重新运行后端

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 提交规范
- 使用清晰的中文提交信息
- 遵循代码规范
- 添加必要的测试
- 更新相关文档

### 开发规范
- 前端使用 ESLint + Prettier
- 后端使用 go fmt
- 提交前运行测试
- 保持代码简洁清晰

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 致谢

感谢以下开源项目的支持：
- [Vue.js](https://vuejs.org/) - 前端框架
- [Go Fiber](https://docs.gofiber.io/) - 后端框架
- [TipTap](https://tiptap.dev/) - 富文本编辑器
- [DaisyUI](https://daisyui.com/) - UI组件库
- [Lucide](https://lucide.dev/) - 图标库

## 联系方式

- **项目地址**: https://github.com/MiXiaoAi/oinote
- **作者**: MiXiaoAi

---

如果这个项目对你有帮助，请给它一个星标！