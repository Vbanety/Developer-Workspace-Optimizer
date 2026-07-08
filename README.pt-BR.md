🇺🇸 [Read in English](README.md)

# 🧹 devopt — Developer Workspace Optimizer

*Saiba o que está consumindo seu disco antes de apagar.*

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![Platform](https://img.shields.io/badge/platform-Linux-333333?logo=linux&logoColor=white)](.claude/contexts/roadmap.md)
[![Status](https://img.shields.io/badge/status-v0.3-success)](CHANGELOG.md)

CLI que escaneia caches de ferramentas de desenvolvimento (npm/yarn/Docker/Gradle/IDEs/browsers/etc), mostra quanto espaço dá pra recuperar e limpa com segurança — nunca sem escanear e confirmar antes.

Não é "mais um cache cleaner": o diferencial é transparência. Antes de apagar qualquer coisa, mostra o que é, quanto ocupa, e por que é seguro remover.

## 😱 Como Começou → 🎉 Onde Chegou

**😱 Começou** do jeito clássico: disco perigosamente cheio, sem ideia de onde foi parar o espaço. A solução óbvia — um script shell encadeando `rm -rf` em algumas pastas de cache — parecia a ferramenta errada pro trabalho. Um path errado, um volume Docker ativo pego no meio, e um script `rm -rf` não limpa a máquina, quebra ela.

Então o pedido de verdade nunca foi *"apaga meu cache"*. Foram umas perguntas que nenhuma ferramenta existente respondia bem:

- 🔍 O que está consumindo meu SSD de verdade?
- ⚠️ O que é seguro apagar, e o que vai quebrar meu ambiente se eu mexer?
- 📊 Quanto eu realmente vou recuperar?
- 🤔 Vale a pena limpar?

Olhando ao redor, a maioria das ferramentas de limpeza de cache é boa na parte de "limpar" e fraca em tudo antes disso — não se explicam, não distinguem artefato de build órfão de trabalho ativo, e raramente consideram como ambiente de dev moderno realmente é: containers Docker, cache de IDE (Cursor/VS Code), toolchains de linguagem, perfis de browser, tudo empilhado.

|  | Cache cleaner típico | **devopt** |
|---|:---:|:---:|
| Explica o que encontrou | ❌ | ✅ |
| Distingue cache de dado real do projeto | ❌ | ✅ |
| Entende Docker / IDEs / toolchains | raramente | ✅ |
| Confirma antes de apagar qualquer coisa | às vezes | ✅ sempre |
| Modo dry-run | raramente | ✅ |
| Mantém log de auditoria | raramente | ✅ |

**🎉 Foi aí que o devopt chegou:** escanear e explicar antes, apagar só o que é comprovadamente seguro, e nunca tocar em projeto, banco de dados ou container rodando. Recuperar espaço em disco não devia significar apostar no seu ambiente de dev.

## 📊 Status

**v0.1, v0.2 e v0.3 — completos.** Linux apenas. Windows e macOS são fases futuras (ver [roadmap](.claude/contexts/roadmap.md)).

| Módulo | O que limpa | Seguro por padrão? |
|---|---|:---:|
| 🧶 `yarn` | Cache de pacotes Yarn | ✅ |
| 📦 `npm` | Cache de pacotes npm | ✅ |
| 📦 `pnpm` | Store content-addressable do pnpm | ✅ |
| 🐘 `gradle` | Cache de build do Gradle | ✅ |
| 🎼 `composer` | Cache de pacotes Composer | ✅ |
| 🎭 `playwright` | Binários de browser do Playwright | ✅ |
| 🤖 `puppeteer` | Binários de browser do Puppeteer | ✅ |
| 🗑️ `trash` | Lixeira da área de trabalho | ✅ |
| 🖱️ `cursor` | Cache do Cursor (não settings/histórico) | ✅ |
| 💻 `vscode` | Cache do VS Code (não settings/histórico) | ✅ |
| 📦 `apt` | Pacotes `.deb` baixados | ⚠️ exige `sudo` |
| 🐳 `docker` | Imagens órfãs, containers parados, build cache — nunca volumes | ⚠️ sempre confirma |
| 📸 `snap` | Revisões desabilitadas (substituídas) de snaps | ⚠️ exige root |

## 📦 Instalação

```sh
go install github.com/Vbanety/Developer-Workspace-Optimizer/cmd/devopt@latest
```

Ou clone e compile localmente:

```sh
git clone https://github.com/Vbanety/Developer-Workspace-Optimizer.git
cd Developer-Workspace-Optimizer
go build -o devopt ./cmd/devopt
```

Requer Go 1.24+.

## ⚡ Uso

```sh
devopt                       # menu interativo (scan → limpeza segura/profunda, escolher módulos, gerar relatório)

devopt report                # só escaneia, nunca apaga nada
devopt report --json         # saída em JSON

devopt clean --dry-run       # mostra o que seria limpo, não toca em nada
devopt clean --module=yarn   # limpa só um módulo (confirma antes)
devopt clean --safe --yes    # só módulos seguros, sem confirmação

devopt log                   # últimas 20 limpezas
devopt log -n 0 --json       # histórico completo em JSON
```

## 🛡️ Nunca Apaga

- 🚫 `~/Documents`, `~/Desktop`, `.config`, `.ssh`.
- 🚫 Diretórios de projeto (detecta `.git`, `package.json`, `go.mod`, etc. na raiz).
- 🚫 Qualquer coisa fora da lista de paths de cada módulo.

Cache menor que 200 MB é ignorado por padrão. Toda limpeza real passa por confirmação a menos que `--yes` seja passado.

## 🧪 Desenvolvimento

```sh
go build ./...
go vet ./...
go test ./...
```

Arquitetura, contrato de módulo e regras de segurança documentados em [`.claude/contexts/`](.claude/contexts/). Pra adicionar um módulo novo, veja `.claude/commands/add-module.md`.

## 📄 Licença

MIT — ver [LICENSE](LICENSE).
