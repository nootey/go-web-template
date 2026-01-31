# Go Web Template

A Go web application template with clean architecture, testing, and CI/CD workflows.

## Quick Start

### Clone the Template
```bash
git clone https://github.com/yourusername/go-web-template.git myapp
cd myapp
```

### Update Module Name

Replace `go-web-template` with your app name:
```bash
# Update go.mod

# Update all imports
```

### Install Dependencies
```bash
go mod download
```

### Setup Environment
```bash
cp .env.example .env
```

### Run Migrations
```bash
make migrate-up
```

### Seed Database
```bash
# Core data only (roles, permissions, root user)
make seed
```

### Run the Application
```bash
make run
```

Visit: `http://127.0.0.1:8080/health`

---

## Architecture

### Adding a New Resource

1. **Create migration:**
```bash
   make migrate-create NAME=create_posts_table
```

2. **Write SQL queries** in `queries/post.sql`:
```sql
   -- name: GetPost :one
   SELECT * FROM posts WHERE id = $1;
   
   -- name: CreatePost :one
   INSERT INTO posts (title, content, user_id)
   VALUES ($1, $2, $3)
   RETURNING *;
```

3. **Generate code:**
```bash
   make sqlc-generate
```

4. **Create service**:
```go
   type PostService struct {
       queries *database.Queries
       logger  *zap.Logger
   }
```

5. **Create handler**:
```go
   func (h *PostHandler) Routes() chi.Router {
       r := chi.NewRouter()
       r.Get("/", h.ListPosts)
       r.Post("/", h.CreatePost)
       return r
   }
```

6. **Wire it up** in `cmd/api/main.go`:
```go
   postService := services.NewPostService(queries, logger)
   postHandler := handlers.NewPostHandler(postService, logger)
   r.Mount("/posts", postHandler.Routes())
```

---

## Testing

### Integration Tests (with real DB)
```bash
go test -v ./internal/services/
```

Uses Testcontainers to spin up PostgreSQL automatically.

### Unit Tests (with mocks)
```bash
go test -v ./internal/handlers/
```

Uses generated mocks to test handlers in isolation.

---

## CI/CD

### GitHub Actions Workflows

The template includes workflows:

- **Lint**:
    - Client and server

- **Tests**:
    - Runs all tests with coverage

**Release**:
    - Auto-creates Git tags and GitHub releases

### Dependabot

- Dependabot is included, for dependency updates.

### Setup Release Automation

-  **Create labels** in GitHub (Settings → Labels):

   | Name | Color | Description |
      |------|-------|-------------|
   | `release:major` | `#d73a4a` | Breaking changes (v1.0.0 → v2.0.0) |
   | `release:minor` | `#fbca04` | New features (v1.0.0 → v1.1.0) |
   | `release:patch` | `#0e8a16` | Bug fixes (v1.0.0 → v1.0.1) |

- **Usage:**
    - Create a PR to `main`
    - Add one of the `release:*` labels
    - Merge the PR
    - Workflow automatically creates tag and release!

---

## Configuration

All configuration is managed through environment variables (`.env`).

Access config anywhere:
```go
import "myapp/internal/config"

cfg := config.Get()
fmt.Println(cfg.Server.Port)
```