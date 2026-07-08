# AGENTS.md

Guidance for ZCode agents working in this repository. For deeper background, also see `CLAUDE.md`, `README.md`, and `dev-docs/`.

## What this project is

**reference** is a local Git-repository reference manager for the AI-assisted coding era. It clones/caches arbitrary remote or local Git repos globally, links them into the current project via Junction/Symlink, and injects AI sub-agent + Skill configs so an AI can read repo source and accumulated Markdown knowledge with zero network latency and zero context pollution.

- Go CLI (`reference`) + Wails v2 desktop GUI (`reference-gui`)
- Module path: `github.com/cicbyte/reference`
- Go **1.25.2** (go-git/v5 needs 1.24+). Pure Go — **no CGO** (SQLite via `github.com/glebarez/sqlite`).

## Layout

| Path | Purpose |
|:---|:---|
| `main.go` | CLI entry; embeds `prompts/**` into `common.PromptsFS`, calls `cmd.Execute()` |
| `cmd/` | Cobra CMD layer. `root.go` is the only root command; subpkgs: `repo/`, `global/`, `proxy/`, `wiki/`, `version/` |
| `internal/logic/` | Business logic, **no cobra dependency**. `repo/`, `global/`, `wiki/` each expose `*Config` + `NewXxxProcessor(...).Execute(ctx)` factories |
| `internal/models/` | GORM models (`Repo`, `AppConfig`, `ProjectSettings`) + settings JSON load/save |
| `internal/utils/` | DB (`GetGormDB` singleton + AutoMigrate), config singleton, paths, formatting |
| `internal/common/` | Global bridges: `AppConfigModel`, `PromptsFS`, asset helpers |
| `internal/log/` | Zap logger + GORM logger + lumberjack rotation |
| `pkg/` | Public Go SDK. `Engine` wraps repo ops; consumers should use this, not `internal/` |
| `prompts/` | Embedded agent templates: `agents/reference-{explorer,analyzer}.{md,toml}`, `skills/reference/SKILL.md` |
| `gui/` | Wails v2 app (Go backend in `app.go`/`binding.go`/`main.go`) + Vue 3 frontend in `gui/frontend/` |
| `scripts/` | `build.py` (cross-compile), `release.py`, `gen_release_notes.py`, `generate_logo.py` |
| `docs/` | Per-command user docs; `dev-docs/` holds internal design + GUI PRD |

## Build, test, run

```bash
go build -o reference.exe .          # dev build (version = "dev")
go test ./...                        # all tests
go test ./internal/logic/repo/...    # focused
go test -run TestParseGitURL_HTTPS ./internal/logic/repo/...
go mod tidy
python scripts/build.py --local      # current platform → dist/ (with ldflags + optional UPX)
```

**Versioning**: `VERSION` file is the single source of truth. CI and `build.py` read it. Version/commit/build-time are injected via `-ldflags -X` into package `github.com/cicbyte/reference/cmd/version`. Don't hardcode versions.

Tests use in-memory SQLite (`file::memory:?cache=shared`) — no external deps. Test style is plain `testing` with a local `assert(t, got, want)` helper (see `url_test.go`); no testify.

## GUI (Wails v2 + Vue 3)

- Backend: `gui/binding.go` binds `ReferenceApp` methods to the frontend — each method delegates to `internal/logic/*` Processors (same layer the CLI uses). Add GUI features by adding methods here.
- Frontend: `gui/frontend/` (Vite + Vue 3 + Ant Design Vue + Pinia + vue-router). Views in `src/views/`, one per CLI area.
- Build: `gui/wails.json` drives Wails (`frontend:build` = `npm run build`, which runs `vue-tsc --noEmit && vite build`).
- Window is **frameless** (`Frameless: true`); `WindowMinimize/Maximize/Close` are no-op stubs in `binding.go` — wire them to `runtime` if implementing window controls.
- `gui/build/` holds `appicon.png` and platform manifests; `gui/frontend/dist/` is `//go:embed`-ed into the binary.

## Architecture rules that matter for edits

1. **CMD/Logic separation is strict.** `cmd/*` binds flags, validates, builds a `*Config`, and calls a Processor's `Execute(ctx)`. Never import cobra from `internal/logic/`. Return structured results, let CMD format output (global `-f table|json|jsonl`).
2. **Adding a command**: create the file under `cmd/<module>/`, add a `NewXxxProcessor` in `internal/logic/<module>/`, then `rootCmd.AddCommand(...)` in `cmd/root.go` `init()`.
3. **Adding a supported AI assistant**: register one entry in `AgentRegistry` (`internal/logic/repo/agent_registry.go`) — inject/doctor/remove all read from this map. Don't scatter agent-specific paths elsewhere.
4. **Two-layer agent design is intentional**: `reference-explorer` (topic Q&A → `<topic>.md`) and `reference-analyzer` (full architecture → `reference.md`, once per repo). Knowledge is written under the wiki dir and Junction-linked into `.reference/wiki/<refName>/`. Preserve this split.
5. **App init order in `cmd/root.go` `init()` is fixed and load-bearing**: `InitAppDirs → LoadConfig → ApplyConfig → InitDataDirs → InitLog → GetGormDB (AutoMigrate) → MigratePathsIfNeeded → EnsureGitInit(wiki) → EnsureGitInit(localwiki)`. Any step failing calls `os.Exit(1)`. Note: `EnsureAutoPull` exists but is **not** called in root `init()`.

## Conventions & gotchas

- **Cross-platform links**: Unix uses Symlink; Windows uses PowerShell `New-Item -ItemType Junction` (no admin rights needed). See `internal/logic/repo/linker.go`.
- **Project data lives in `.reference/`** (Junctions to global cache + wiki). `.reference/`, `.zcode/`, `.claude/`, etc. are gitignored — they are per-project generated state, do not commit them.
- **User data lives in `~/.cicbyte/reference/`**: `config/config.yaml`, `db/app.db`, `repos/`, `wiki/` (remote, nested `<platform>/<namespace>/<repo>/`), `localwiki/` (local repos), `logs/`.
- **`repo add` default changed**: cached repos are NOT auto-updated; require explicit `--update`/`-u`.
- **Remote vs local wiki are separate Git repos** (`wiki/` and `localwiki/`); switch target with `--local`/`-l`.
- **`ProjectSettings.Agent` (single) is deprecated** → migrated automatically to `Agents` array. Always write the array form.
- Logging is Zap (`log.Info`, `log.Error`, with `zap.String`/`zap.Error` fields); never `fmt.Println` from logic layer (only from CMD/user-facing output).
- Global state is bridged through `common.AppConfigModel` and the `utils.ConfigInstance` singleton — read config via these, not by re-parsing `config.yaml`.
- Commit messages follow Conventional Commits in Chinese (e.g. `feat(gui): ...`, `refactor(repo/cache): ...`); see `git log` for tone.

## Docs to read before touching sensitive areas

- `docs/agent-platform-adapter.md` — how each AI platform's config dir/files are targeted
- `dev-docs/reference-gui-prd.md` — GUI scope, views, and binding contracts
- `dev-docs/global-management-backend.md` — cross-project GC/stats logic
- `CHANGELOG.md` — recent behavior changes (e.g. `--update` default, multi-agent migration)
