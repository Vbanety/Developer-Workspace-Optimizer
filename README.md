🇧🇷 [Leia em português](README.pt-BR.md)

# devopt — Developer Workspace Optimizer

A CLI that scans development-tool caches (npm/yarn/Docker/Gradle/IDEs/browsers/etc), shows how much disk space you can recover, and cleans it safely — never without scanning and confirming first.

Not "just another cache cleaner": the whole point is transparency. Before deleting anything, it shows you what it is, how much space it takes, and why it's safe to remove.

## Why this exists

This started the way it usually does: a disk alarmingly close to full, and no idea where the space went. The obvious fix — a shell script chaining `rm -rf` on a few cache folders — felt like the wrong tool for the job. One wrong path, one active Docker volume caught in the blast radius, and a `rm -rf` script doesn't clean your machine, it breaks it.

So the real ask was never "delete my caches." It was a handful of questions no existing tool actually answered well:

- What's actually eating my SSD?
- What's safe to delete, and what will break my setup if I touch it?
- How much will I actually get back?
- Is it even worth cleaning?

Looking around, most cache-cleaning tools are good at the "clean" part and weak on everything before it — they don't explain themselves, don't distinguish orphaned build artifacts from your active work, and rarely account for how modern dev environments actually look: Docker containers, IDE caches (Cursor/VS Code), language toolchains, browser profiles, all layered on top of each other.

devopt is built around the part that was missing: **scan and explain first, delete only what's proven safe, and never touch a project, a database, or a running container.** Cleaning disk space shouldn't require gambling with your dev environment.

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
