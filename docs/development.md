# Development

## Database

```sh
make migrate-up                          # run all migrations
make migrate-down                        # roll back the last migration
make migrate-create NAME=create_x_table  # add a new migration
make migrate-fresh                       # drop, migrate and seed
make seed                                # core data: roles, permissions, root user
```

Queries live in `internal/queries/`. Run `make sqlc` after you change them.

## Adding a new resource

Each resource is a domain in `internal/domains/<name>/`, with a `handler.go`, `service.go` and `models.go`.
`internal/domains/user` is the smallest example.

1. Add a migration with `make migrate-create NAME=create_posts_table`, and write the table in it.
2. Add the queries to `internal/queries/post.sql`, then run `make sqlc`:

    ```sql
    -- name: GetPost :one
    SELECT * FROM posts WHERE id = $1;
    ```

3. Add the service. It takes the generated queries, and the handler uses it through an interface:

    ```go
    type PostServiceInterface interface {
        GetPost(ctx context.Context, id int64) (*Post, error)
    }

    type PostService struct {
        queries *database.Queries
    }
    ```

4. Add the handler and its routes:

    ```go
    func (h *PostHandler) Routes() chi.Router {
        r := chi.NewRouter()
        r.Get("/{id}", h.GetPost)
        return r
    }
    ```

5. Wire it up in the API entry point. Build the service and the handler, then mount the handler routes with the
   public or the protected routes:

    ```go
    postService := post.NewPostService(queries)
    postHandler := post.NewPostHandler(postService, responder)
    ```

    ```go
    r.Mount("/posts", h.Post.Routes())
    ```

6. Add the interface to `.mockery.yaml` and run `make mock`, so the handler tests can use a mock service.

## Configuration

All config comes from environment variables, read from `.env`. You can set each config key with an env var: the key
in upper case, with dots replaced by underscores (`server.port` is `SERVER_PORT`).

```go
cfg := config.Get()
fmt.Println(cfg.Server.Port)
```

## Testing

```sh
make test              # all tests
make test-unit         # unit tests, with mocks
make test-integration  # integration tests, with a real database
```

Integration tests use Testcontainers to start PostgreSQL. Docker must be running.
Run `make mock` after you change an interface.

## Linting

```sh
make lint                      # server
cd client && pnpm run lint     # client
```

`make pre-push` runs the full check: mocks, build, lint and tests for the server and the client.

## CI/CD

GitHub Actions runs these workflows:

- **Lint** - client and server
- **Tests** - all server tests with coverage
- **Build Docker Image** - builds the API and client images for amd64 and arm64 on each push to `main` or `develop`
- **Release** - creates Git tags and GitHub releases, then pushes the images to Docker Hub

### Release automation

Create these labels in GitHub (Settings → Labels):

| Name | Color | Description |
|------|-------|-------------|
| `release:minor` | `#fbca04` | New features (v1.0.0 → v1.1.0) |
| `release:fix` | `#0e8a16` | Bug fixes (v1.0.0 → v1.0.1) |

Add these secrets in GitHub (Settings → Secrets and variables → Actions), so a release can push the images:

- `DOCKER_HUB_USERNAME`
- `DOCKER_HUB_ACCESS_TOKEN`

To make a release:

1. Create a PR to `main`
2. Add one of the `release:*` labels
3. Merge the PR

The workflow then creates the tag and the release, and pushes the images with the version tag and `latest`.
