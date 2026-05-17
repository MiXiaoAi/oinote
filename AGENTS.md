# oinote

Collaborative note-taking app (like Notion) with channels, real-time collab, and AI features.

## Structure

- `frontend/` - Vue 3 + Vite + Tailwind CSS + Tiptap + Pinia
- `backend/` - Go + Fiber + GORM + SQLite

## Commands

```bash
# Frontend
cd frontend && npm run dev      # dev server: http://localhost:5173

# Backend
cd backend && go run main.go  # or ./backend/oinote.exe
```

Both run simultaneously. Backend API at `http://localhost:3000`.

## Database

SQLite at `backend/data/oinote.db`. Created automatically on first run.

**Default admin**: `admin` / `admin` (auto-created if no users exist).

## Important Quirks

- Backend body limit: 2GB (`configBodyLimit` in main.go:38)
- Search route must be BEFORE `notes/:id` route (line 124 in main.go)
- `/api/notes/search` vs `/api/notes/:id` route ordering matters
- WebSocket endpoints: `/ws` (notifications) and `/ws/collab` (Yjs sync)
- Frontend uses Tailwind v4 with `@tailwindcss/vite` plugin (not postcss config)
- CORS allows all origins (`AllowOrigins: "*"`)

## What's NOT configured

No linter, formatter, typechecker, CI, or `.gitignore` exists. There are no quality gates to run after changes. No lockfile — `npm install` results are non-deterministic.

## Key Files

- `backend/main.go` — all route registration and server setup
- `backend/config/db.go` — SQLite connection, GORM auto-migration, default admin seeding
- `backend/internal/middleware/auth.go` — JWT auth (secret hardcoded: `oinote_secret_key_123456`)
- `backend/internal/models/schema.go` — all GORM models (User, Channel, Note, etc.)
- `frontend/src/utils/urlHelper.js` — API base URL resolution (always `localhost:3000` in dev)
- `frontend/src/style.css` — Tailwind v4 + DaisyUI theme config with custom OKLCH colors

## Tech Stack Notes

- **Tailwind v4**: uses `@import "tailwindcss"` and `@plugin "daisyui"` syntax, not v3 `@tailwind` directives
- **DaisyUI v5**: component classes available (btn, card, modal, etc.) — theme uses custom grayscale OKLCH palette
- **Tiptap**: rich-text editor, split into manual chunk in Vite config
- **Go module**: `github.com/MiXiaoAi/oinote/backend`, requires Go 1.25.6
- **SQLite driver**: `glebarez/sqlite` (pure Go, no CGO), not `mattn/go-sqlite3`
