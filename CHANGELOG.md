# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [Unreleased] — v0.1.0

### Adicionado
- Scaffold do projeto: `internal/core` (contrato `Module`, safety guard, registry), `internal/config` (paths default por OS), `internal/report` (renderização de tabela).
- Comandos `devopt report` (`--json`) e `devopt clean` (`--safe`, `--deep`, `--module`, `--dry-run`, `--yes`).
- Módulo genérico `DirCache` (Detect/Calculate/Clean completos) cobrindo os 9 módulos de v0.1: `yarn`, `npm`, `pnpm`, `gradle`, `composer`, `playwright`, `puppeteer`, `trash` (safe) e `apt` (não-safe, path de sistema).
- `Calculate` pula subpastas sem permissão de leitura em vez de abortar o scan (relevante pro `apt`, cujo `partial/` é root-only).
- Never-delete guard: `~/Documents`, `~/Desktop`, `.config`, `.ssh`, diretórios de projeto.
- Threshold de 200 MB pra pular caches pequenos automaticamente.
