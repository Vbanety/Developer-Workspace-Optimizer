# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [Unreleased]

### Adicionado — v0.3
- `internal/audit`: log de auditoria em JSONL (`~/.local/share/devopt/history.jsonl`) — todo `Clean()` real (CLI e TUI) grava módulo, path, bytes liberados, dry-run, skip/erro. `Record`/`Read` best-effort (falha ao logar não derruba a limpeza).
- Comando `devopt log [-n N] [--json]` — lê o histórico (default: últimas 20).
- `audit.OverrideLogPathForTest`, usado pelos testes que exercitam `Clean()` de verdade (`internal/tui`) pra não escrever no histórico real do usuário durante `go test`.

### Adicionado — v0.2
- Módulo `docker` (bespoke): soma `Reclaimable` de `Images`/`Containers`/`Build Cache` via `docker system df --format json`; limpa via `docker system prune -f` (nunca `--volumes`, nunca `--all`). `Local Volumes` nunca é somado nem tocado. `Safe() == false` sempre.
- Módulo `snap` (bespoke): parseia `snap list --all` (sem suporte a JSON) pra achar revisões `disabled`, soma tamanho via `os.Stat` no squashfs de cada uma. Limpa via `snap remove --revision=<rev>` por item — uma falha não bloqueia as demais. `Safe() == false` sempre.
- Módulo genérico `MultiDirCache` cobrindo `cursor` e `vscode`: soma/limpa só as subpastas de cache puro curadas em `ElectronCacheSubdirs` (Cache, GPUCache, CachedExtensionVSIXs, etc.), nunca a raiz inteira (`User/`, `snapshots`, sessão/login preservados).
- Helpers `dirSize`/`emptyDir` extraídos de `DirCache` pra serem reusados por `MultiDirCache` (evita duplicar a lógica de walk).
- `report.Scan` extraído de `cmd/devopt` pra ser compartilhado com o menu interativo.
- Menu interativo (`internal/tui`, Bubble Tea): scan → Limpeza segura/profunda/Escolher módulos/Gerar relatório/Sair → confirmação → execução → relatório final (com espaço livre antes/depois). `devopt` sem subcomando abre o menu; `report`/`clean` continuam para scripts/CI.

### Adicionado — v0.1.0
- Scaffold do projeto: `internal/core` (contrato `Module`, safety guard, registry), `internal/config` (paths default por OS), `internal/report` (renderização de tabela).
- Comandos `devopt report` (`--json`) e `devopt clean` (`--safe`, `--deep`, `--module`, `--dry-run`, `--yes`).
- Módulo genérico `DirCache` (Detect/Calculate/Clean completos) cobrindo os 9 módulos de v0.1: `yarn`, `npm`, `pnpm`, `gradle`, `composer`, `playwright`, `puppeteer`, `trash` (safe) e `apt` (não-safe, path de sistema).
- `Calculate` pula subpastas sem permissão de leitura em vez de abortar o scan (relevante pro `apt`, cujo `partial/` é root-only).
- Never-delete guard: `~/Documents`, `~/Desktop`, `.config`, `.ssh`, diretórios de projeto.
- Threshold de 200 MB pra pular caches pequenos automaticamente.
