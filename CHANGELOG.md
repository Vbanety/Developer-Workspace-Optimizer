# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [Unreleased] — v0.1.0

### Adicionado
- Scaffold do projeto: `internal/core` (contrato `Module`, safety guard, registry), `internal/config` (paths default por OS), `internal/report` (renderização de tabela).
- Comandos `devopt report` (`--json`) e `devopt clean` (`--safe`, `--deep`, `--module`, `--dry-run`, `--yes`).
- Módulo `yarn` implementado (Detect/Calculate/Clean completos).
- Módulos stub (`npm`, `pnpm`, `gradle`, `composer`, `playwright`, `puppeteer`, `apt`, `trash`): detecção real, limpeza pendente.
- Never-delete guard: `~/Documents`, `~/Desktop`, `.config`, `.ssh`, diretórios de projeto.
- Threshold de 200 MB pra pular caches pequenos automaticamente.
