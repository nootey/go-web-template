# Go Web Template

A Go web application template with clean architecture, testing, and CI/CD workflows.
It comes with a Vue client, auth with sessions, PostgreSQL, Redis, an observability stack and Docker deploys.

## Quick Start

### Requirements

- Go >= 1.27
- Node >= 22.12 and pnpm
- PostgreSQL
- Redis
- Docker (`make run` starts the observability stack with Docker Compose)

### Clone the template

```sh
git clone https://github.com/yourusername/go-web-template.git myapp
cd myapp
```

### Rename the app

Replace `go-web-template` with your app name in the Go module, the imports and the config files:

```sh
grep -rlI 'go-web-template' --exclude-dir=.git --exclude-dir=node_modules . | xargs sed -i 's/go-web-template/myapp/g'
```

This also renames the Docker images in `docker-compose.yaml` and `.github/workflows/build-image.yml`.
Check that the image names point to your Docker Hub account.
The client shows the name `MyApp` in `client/index.html`, `client/src/pkg/router/router.ts` and
`client/src/pkg/components/AuthSkeleton.vue`.

### Configure

Create the configuration files from their examples and edit them:

- `./.env.example` -> `./.env`
- `./client/.env.example` -> `./client/.env` (optional, defaults work for local dev)

### Run

Run the migrations and seeders, then start the server and the client:

```sh
make migrate-up
make seed
make run
cd client && pnpm install && pnpm run dev
```

By default, the app is available at http://localhost:3030, and the API health check at http://localhost:8080/health.
Log in with the root user from `.env` (`SEED_ROOT_USER` and `SEED_ROOT_PASSWORD`).
Sign up is off by default. Set `APP_ALLOW_SIGNUP=true` to turn it on.

## Docs

- [docs/development.md](./docs/development.md) - database, adding a resource, config, tests, linting and releases
- [docs/docker.md](./docs/docker.md) - run the app with Docker on a server or an ARM board
