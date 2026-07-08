🇧🇷 [Leia em português](README.pt-BR.md)

# 🧹 devopt — Developer Workspace Optimizer

*Know what's eating your disk before you delete it.*

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![Platform](https://img.shields.io/badge/platform-Linux-333333?logo=linux&logoColor=white)](.claude/contexts/roadmap.md)
[![Status](https://img.shields.io/badge/status-v0.3-success)](CHANGELOG.md)

A CLI that scans development-tool caches (npm/yarn/Docker/Gradle/IDEs/browsers/etc), shows how much disk space you can recover, and cleans it safely — never without scanning and confirming first.

Not "just another cache cleaner": the whole point is transparency. Before deleting anything, it shows you what it is, how much space it takes, and why it's safe to remove.

## 😱 How It Started → 🎉 How It's Going

**😱 It started** the way it usually does: a disk alarmingly close to full, and no idea where the space went. The obvious fix — a shell script chaining `rm -rf` on a few cache folders — felt like the wrong tool for the job. One wrong path, one active Docker volume caught in the blast radius, and a `rm -rf` script doesn't clean your machine, it breaks it.

So the real ask was never *"delete my caches."* It was a handful of questions no existing tool actually answered well:

- 🔍 What's actually eating my SSD?
- ⚠️ What's safe to delete, and what will break my setup if I touch it?
- 📊 How much will I actually get back?
- 🤔 Is it even worth cleaning?

Looking around, most cache-cleaning tools are good at the "clean" part and weak on everything before it — they don't explain themselves, don't distinguish orphaned build artifacts from your active work, and rarely account for how modern dev environments actually look: Docker containers, IDE caches (Cursor/VS Code), language toolchains, browser profiles, all layered on top of each other.

|  | Typical cache cleaner | **devopt** |
|---|:---:|:---:|
| Explains what it found | ❌ | ✅ |
| Tells cache apart from real project data | ❌ | ✅ |
| Understands Docker / IDEs / toolchains | rarely | ✅ |
| Confirms before deleting anything | sometimes | ✅ always |
| Dry-run mode | rarely | ✅ |
| Keeps an audit log | rarely | ✅ |

**🎉 That's where devopt landed:** scan and explain first, delete only what's proven safe, and never touch a project, a database, or a running container. Recovering disk space shouldn't mean gambling with your dev environment.

## 📊 Status

**v0.1, v0.2, and v0.3 — complete.** Linux only for now; Windows and macOS are future phases (see [roadmap](.claude/contexts/roadmap.md)).

| Module | What it cleans | Safe by default? |
|---|---|:---:|
| 🧶 `yarn` | Yarn package cache | ✅ |
| 📦 `npm` | npm package cache | ✅ |
| 📦 `pnpm` | pnpm content-addressable store | ✅ |
| 🐘 `gradle` | Gradle build cache | ✅ |
| 🎼 `composer` | Composer package cache | ✅ |
| 🎭 `playwright` | Playwright browser binaries | ✅ |
| 🤖 `puppeteer` | Puppeteer browser binaries | ✅ |
| 🗑️ `trash` | Desktop trash/recycle bin | ✅ |
| 🖱️ `cursor` | Cursor IDE cache (not settings/history) | ✅ |
| 💻 `vscode` | VS Code cache (not settings/history) | ✅ |
| 📦 `apt` | Downloaded `.deb` packages | ⚠️ needs `sudo` |
| 🐳 `docker` | Dangling images, stopped containers, build cache — never volumes | ⚠️ always confirms |
| 📸 `snap` | Disabled (superseded) snap revisions | ⚠️ needs root |

## 📦 Install

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

## ⚡ Usage

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

## 🛡️ Never Deletes

- 🚫 `~/Documents`, `~/Desktop`, `.config`, `.ssh`.
- 🚫 Project directories (detects `.git`, `package.json`, `go.mod`, etc. at the root).
- 🚫 Anything outside each module's own path list.

Caches smaller than 200 MB are skipped by default. Every real cleanup asks for confirmation unless `--yes` is passed.

## 🧪 Development

```sh
go build ./...
go vet ./...
go test ./...
```

Architecture, the module contract, and safety rules are documented in [`.claude/contexts/`](.claude/contexts/). To add a new module, see `.claude/commands/add-module.md`.

## 📄 License

MIT — see [LICENSE](LICENSE).
