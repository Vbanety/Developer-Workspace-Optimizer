🇺🇸 [Read in English](README.md)

# devopt — Developer Workspace Optimizer

CLI que escaneia caches de ferramentas de desenvolvimento (npm/yarn/Docker/Gradle/IDEs/browsers/etc), mostra quanto espaço dá pra recuperar e limpa com segurança — nunca sem escanear e confirmar antes.

Não é "mais um cache cleaner": o diferencial é transparência. Antes de apagar qualquer coisa, mostra o que é, quanto ocupa, e por que é seguro remover.

## Por que existe

Começou do jeito clássico: disco perigosamente cheio, sem ideia de onde foi parar o espaço. A solução óbvia — um script shell encadeando `rm -rf` em algumas pastas de cache — parecia a ferramenta errada pro trabalho. Um path errado, um volume Docker ativo pego no meio, e um script `rm -rf` não limpa a máquina, quebra ela.

Então o pedido de verdade nunca foi "apaga meu cache". Foram umas perguntas que nenhuma ferramenta existente respondia bem:

- O que está consumindo meu SSD de verdade?
- O que é seguro apagar, e o que vai quebrar meu ambiente se eu mexer?
- Quanto eu realmente vou recuperar?
- Vale a pena limpar?

Olhando ao redor, a maioria das ferramentas de limpeza de cache é boa na parte de "limpar" e fraca em tudo antes disso — não se explicam, não distinguem artefato de build órfão de trabalho ativo, e raramente consideram como ambiente de dev moderno realmente é: containers Docker, cache de IDE (Cursor/VS Code), toolchains de linguagem, perfis de browser, tudo empilhado.

O devopt nasce em torno da parte que faltava: **escanear e explicar antes, apagar só o que é comprovadamente seguro, e nunca tocar em projeto, banco de dados ou container rodando.** Recuperar espaço em disco não devia significar apostar no seu ambiente de dev.

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
