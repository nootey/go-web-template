# Docker

This guide helps you set up and run the app with Docker Compose.

## Configure your environment

All config lives in one `.env` file next to `docker-compose.yaml`. Copy `.env.example` to `.env` and edit it.
Compose reads it, and the app containers get all of it. Inside Docker, Compose sets only the container hosts and ports.

These are the values to set for a deployment:

| Variable | Default | Description |
|----------|---------|-------------|
| `WEB_CLIENT_DOMAIN` | `localhost` | Your domain. Traefik routes it, and email links use it. |
| `APP_ENVIRONMENT` | `local` | `production` on a server with Traefik. `development` on an ARM board, which has no HTTPS. |
| `TRAEFIK_EMAIL` | `support@example.com` | Email for Let's Encrypt. |
| `DATABASE_USER` | `postgres` | Database user. |
| `DATABASE_PASSWORD` | `postgres` | Database password. |
| `DATABASE_NAME` | `go-web-template` | Database name. |
| `REDIS_PASSWORD` | empty | Redis password. Required in production. |
| `SEED_ROOT_USER` | `root@local.host` | Email of the first user. |
| `SEED_ROOT_PASSWORD` | `password` | Password of the first user. |
| `MAILER_HOST` | empty | SMTP server for email links. Empty writes the links to the API log. |
| `GRAFANA_PASSWORD` | `admin` | Grafana admin password. |
| `WEB_CLIENT_PORT` | `5173` | ARM only: the host port of the app. |
| `GRAFANA_PORT` | `3000` | ARM only: the host port of Grafana. |

> ⚠️ **Warning:** The defaults are not safe. Set your own passwords before you deploy publicly.

## Images

`docker-compose.yaml` pulls the published images, `nootey/go-web-template` and `nootey/go-web-template-client`.
Each release builds them for amd64 and arm64. `docker-compose.build.yaml` builds them from source instead.

## Run on a server

Traefik gets a TLS certificate from Let's Encrypt, so your domain must point to the server, and ports 80 and 443 must be open.
These commands build the images from source.

For the first time setup, you must run the migrations and seed the root user:

```sh
make docker-migrate
make docker-seed
```

To start all services, including the observability stack, use:

```sh
make docker-up
```

The app is then available at `https://<your-domain>`.
Grafana is available at `https://grafana.<your-domain>`.

To restart or stop all services, use:

```sh
make docker-restart
make docker-down
```

## Run on an ARM board

For a Raspberry Pi or another ARM board. The board runs no Traefik.
The client serves the app on `WEB_CLIENT_PORT`, and Grafana runs on `GRAFANA_PORT`.
Set `APP_ENVIRONMENT=development`: production needs HTTPS for its cookies.

```sh
make docker-arm-migrate
make docker-arm-seed
make docker-arm-up
```

The app is then available at `http://<WEB_CLIENT_DOMAIN>:<WEB_CLIENT_PORT>`.
Each command pulls the latest published images, so `make docker-arm-up` also updates the app.
To build on the board instead, add `build=1`, for example `make docker-arm-up build=1`.

To restart or stop all services, use:

```sh
make docker-arm-restart
make docker-arm-down
```
