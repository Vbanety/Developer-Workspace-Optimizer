🇧🇷 [Leia em português](README.pt-BR.md)

# devopt — Developer Workspace Optimizer

A CLI that scans development-tool caches (npm/yarn/Docker/Gradle/IDEs/browsers/etc), shows how much disk space you can recover, and cleans it safely — never without scanning and confirming first.

Not "just another cache cleaner": the whole point is transparency. Before deleting anything, it shows you what it is, how much space it takes, and why it's safe to remove.

## Status

**v0.1, v0.2, and v0.3 — complete.** Linux only for now; Windows and macOS are future phases (see [roadmap](.claude/contexts/roadmap.md)).

Modules: `yarn`, `npm`, `pnpm`, `gradle`, `composer`, `playwright`, `puppeteer`, `trash`, `cursor`, `vscode` (safe, no extra confirmation needed); `apt` (system path — usually needs `sudo` to actually remove), `docker` (containers/images — never `Safe()`, always confirms; never touches volumes), and `snap` (disabled revisions — never `Safe()`, removal needs root).

## Install

```sh
go install github.com/Vbanety/Developer-Workspace-Optimizer/cmd/devopt@latest
```

Or clone and build locally:

```sh
git clone https://github.com/Vbanety/Developer-Workspace-Optimizer.git
cd Developer-Workspace-Optimizer
go build -o devopt ./cmd/devopt
```

Requires Go 1.24+.

## Usage

```sh
devopt                       # interactive menu (scan → safe/deep clean, pick modules, write report)

devopt report                # scan only, never deletes anything
devopt report --json         # JSON output

devopt clean --dry-run       # show what would be cleaned, touch nothing
devopt clean --module=yarn   # clean a single module (confirms first)
devopt clean --safe --yes    # only modules marked safe, no confirmation prompt

devopt log                   # last 20 cleanup actions
devopt log -n 0 --json       # full audit history as JSON
```

## Never deletes

- `~/Documents`, `~/Desktop`, `.config`, `.ssh`.
- Project directories (detects `.git`, `package.json`, `go.mod`, etc. at the root).
- Anything outside each module's own path list.

Caches smaller than 200 MB are skipped by default. Every real cleanup asks for confirmation unless `--yes` is passed.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

Architecture, the module contract, and safety rules are documented in [`.claude/contexts/`](.claude/contexts/). To add a new module, see `.claude/commands/add-module.md`.

## License

MIT — see [LICENSE](LICENSE).
