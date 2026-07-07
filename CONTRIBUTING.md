# Contribuindo com o devopt

## Antes de tudo

Leia `.claude/contexts/architecture.md` (contrato `Module`) e `.claude/contexts/safety.md` (regras de segurança). Toda contribuição que mexe em `Clean()` precisa respeitar o never-delete guard e o threshold de 200 MB — sem exceção.

## Adicionando um módulo novo

Siga `.claude/commands/add-module.md`. Resumo:

1. Confira se já não existe módulo igual em `internal/modules/<os>/`.
2. Use `internal/modules/linux/yarn.go` como referência de implementação completa.
3. Adicione o path default em `internal/config/rules.go`.
4. `Safe() == true` só se for cache puro e comprovadamente órfão.
5. Registre em `cmd/devopt/main.go` (`buildRegistry`).
6. Escreva teste cobrindo `Calculate` e `Clean(dryRun=true)` num fixture temporário.

## Rodando localmente

```sh
go build ./...
go vet ./...
go test ./...
```

## Pull requests

- Um módulo/feature por PR.
- Rode `go vet`/`go test` antes de abrir.
- Se a mudança afeta segurança (guard, threshold, paths never-delete), explique o raciocínio na descrição do PR.
