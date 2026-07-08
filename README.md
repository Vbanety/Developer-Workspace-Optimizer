# devopt — Developer Workspace Optimizer

CLI que escaneia caches de ferramentas de desenvolvimento (npm/yarn/Docker/Gradle/IDEs/browsers/etc), mostra quanto espaço dá pra recuperar e limpa com segurança — nunca sem escanear e confirmar antes.

Não é "mais um cache cleaner": o diferencial é transparência. Antes de apagar qualquer coisa, mostra o que é, quanto ocupa, e por que é seguro remover.

## Status

**v0.1, v0.2 e v0.3 — completos.** Linux apenas. Windows e macOS são fases futuras (ver [roadmap](.claude/contexts/roadmap.md)).

Módulos: `yarn`, `npm`, `pnpm`, `gradle`, `composer`, `playwright`, `puppeteer`, `trash`, `cursor`, `vscode` (seguros, sem confirmação extra); `apt` (path de sistema — normalmente exige `sudo` pra remover de verdade), `docker` (containers/imagens — nunca `Safe()`, sempre confirma; nunca mexe em volumes) e `snap` (revisões desabilitadas — nunca `Safe()`, remoção exige root).

## Uso

```sh
go run ./cmd/devopt              # menu interativo (scan → limpeza segura/profunda/escolher módulos/relatório)

go run ./cmd/devopt report              # só escaneia, nunca apaga nada
go run ./cmd/devopt report --json       # saída em JSON

go run ./cmd/devopt clean --dry-run     # mostra o que seria limpo
go run ./cmd/devopt clean --module=yarn # limpa só um módulo (confirma antes)
go run ./cmd/devopt clean --safe --yes  # só módulos seguros, sem confirmação

go run ./cmd/devopt log                 # histórico das últimas 20 limpezas
go run ./cmd/devopt log -n 0 --json     # histórico completo em JSON
```

## Nunca apaga

- `~/Documents`, `~/Desktop`, `.config`, `.ssh`.
- Diretórios de projeto (detecta `.git`, `package.json`, `go.mod`, etc. na raiz).
- Qualquer coisa fora da lista de paths de cada módulo.

Cache menor que 200 MB é ignorado por padrão. Toda limpeza real passa por confirmação a menos que `--yes` seja passado.

## Desenvolvimento

```sh
go build ./...
go vet ./...
go test ./...
```

Arquitetura, contrato de módulo e regras de segurança documentados em [`.claude/contexts/`](.claude/contexts/). Pra adicionar um módulo novo, veja `.claude/commands/add-module.md`.

## Licença

MIT — ver [LICENSE](LICENSE).
