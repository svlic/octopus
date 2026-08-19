# AGENTS.md — Octopus

LLM API aggregation gateway. Go (Gin + GORM + Cobra) backend with an embedded Vite static admin UI.

Module: `github.com/bestruirui/octopus` · Go **1.26.4** · Node 18+ · **pnpm** (not npm/yarn)

## Commands

```bash
# Backend only (needs embedded UI already present under static/out/)
go run main.go start
go run main.go start --config /path/to/config.json

# Frontend production build → Vite writes directly to static/out (embed path)
cd web && pnpm install && pnpm run build && cd ..

# Full release build (frontend + multi-arch Go binaries + archives)
bash scripts/build.sh release   # needs go, pnpm, python3, git, curl, zip/tar, md5sum

# Tests (sparse; no repo-wide CI test job)
go test ./...
go test ./internal/op/ -run TestName -count=1
go test ./internal/relay/ -count=1

# Frontend dev (separate process; do not rely on embed)
cd web && pnpm install
VITE_API_BASE_URL="http://127.0.0.1:8080" pnpm run dev   # :3000
# other terminal:
go run main.go start                                            # :8080

cd web && pnpm run lint
```

There is **no** Makefile, golangci-lint config, or pre-commit hook. CI (`.github/workflows/release.yaml`) only runs `bash scripts/build.sh release` on push to `master`.

## Critical: static embed

- `static/static.go` does `//go:embed all:out` → serves `static/out`.
- `static/out/*` is generated and gitignored. A clean clone **cannot** `go build` / `go run` until you build the web app into `static/out`.
- Vite writes directly to `static/out` (`web/vite.config.ts`).
- Admin API base defaults to relative `"."` (`web/src/api/client.ts`). Only set `VITE_API_BASE_URL` for the separate Vite dev server.

## Layout (where to edit)

| Path | Role |
|------|------|
| `main.go`, `cmd/` | CLI entry (`start`, `version`) |
| `internal/server/` | Gin server, middleware, `handlers/`, route registry |
| `internal/relay/` | AxonHub-based proxy, protocol conversion, retries, logs, metrics |
| `internal/op/` | Business ops + in-memory caches (handlers call here, not DB directly) |
| `internal/model/` | GORM models |
| `internal/db/`, `internal/db/migrate/` | DB init; numbered migrations via `init()` |
| `internal/task/` | Background jobs (price sync, stats flush, model sync, base-url delay) |
| `internal/price/`, `scripts/updatePrice.py` | Pricing (models.dev); Python used in release pipeline |
| `internal/conf/` | Viper config, banner, version ldflags, `IsDebug` |
| `web/` | Vite React admin UI under `web/src` |
| `static/` | Embed root only — do not hand-edit generated `out/` |
| `scripts/build.sh`, `scripts/dockerfile/` | Release + Docker images |

### Request paths

- **Public LLM API** (API key): `/v1/chat/completions`, `/v1/responses`, `/v1/messages`, `/v1/embeddings`, `/v1/images/*`, `/v1/models`
- **Admin API** (JWT Bearer): `/api/v1/*` (user, channel, group, model, apikey, setting, stats, log)
- Flow: handler → `relay` → AxonHub inbound/outbound pipeline → active group item → upstream
- **Group name** is the external `model` clients send (not the upstream model id alone)

### Route registration quirk

Handlers register routes in `func init()` via `router.NewGroupRouter`. `server.Start` blank-imports `_ "…/internal/server/handlers"` so all handler files must stay in package `handlers`. Adding a new HTTP surface = new/extended file under `handlers/` with `init()` registration — not a central route table.

## Config & runtime

- Default config: `data/config.json` (auto-created). Override: `--config` or env.
- Env prefix: `OCTOPUS_` + path with `_` (e.g. `OCTOPUS_SERVER_PORT`, `OCTOPUS_DATABASE_TYPE`, `OCTOPUS_DATABASE_PATH`).
- Debug: `OCTOPUS_DEBUG=true` → Gin debug + GORM SQL logs + request logger middleware.
- DB: `sqlite` (default `data/data.db`), `mysql`, `postgres`/`postgresql`. MySQL/Postgres DB must exist; tables migrate on start (`BeforeAutoMigrate` → `AutoMigrate` models → `AfterAutoMigrate`).
- First login defaults: `admin` / `admin` (change immediately).
- Stats and relay logs buffer in memory and flush on an interval (`internal/task`). Prefer graceful shutdown (`SIGINT`/`SIGTERM`); `kill -9` drops unflushed stats.
- Working data dirs created at runtime: `data/`.

## Conventions

- **Push target**: the `dyna` branch pushes to `git@github.com:svlic/octopus.git` (a separate remote/fork), not `origin` (`github.com/bestruirui/octopus.git`). Add it once with `git remote add svlic git@github.com:svlic/octopus.git` if missing, then `git push svlic dyna`.
- PRs: **one theme per PR** (`CONTRIBUTING.md`). AI assist OK; human review required before submit.
- Prefer extending existing `op` + handler patterns over new top-level packages.
- Transformer work: register in `internal/relay/transformers.go` and keep supported channel providers in sync.
- Do not commit `data/`, `build/`, or generated `static/out/*`.
- Version/ldflags wired in `scripts/build.sh` → `internal/conf.Version|BuildTime|Author|Commit`.

## Docs

- Product/ops detail: `README.md` / `README_zh.md` (channel base URLs, LB modes, client examples).
- Trust executable sources (`go.mod`, `scripts/build.sh`, handler `init()`s) over prose when they disagree.
