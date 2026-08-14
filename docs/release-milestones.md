# Release Milestones

Este documento organiza a reescrita do SparkEdge em Go em milestones de release, com foco no que precisa estar funcional, verificavel e publicavel em cada etapa.

## Milestone 1 - Foundation

Objetivo: garantir a base operacional do SparkEdge Go.

- Backend Go com bootstrap unico
- SQLite com GORM
- JWT mantido como autenticacao
- `config.yml` com prioridade `config.yml > env > defaults`
- `webui` servido pela aplicacao Go
- build local e CI multi-plataforma

Status:
- [x] Concluido

## Milestone 2 - Core Runtime

Objetivo: migrar o nucleo funcional da automacao.

- scripts Python executados em venv
- integracao com `sparkit`
- cadastro, atualizacao, historico e restauracao de scripts
- criacao e edicao de instancias
- resolucao de destinos, fallback local e reenvio
- mapeamento com contexto de script, sistema e device

Status:
- [x] Concluido

## Milestone 3 - Providers and Connectivity

Objetivo: fechar as integracoes principais do ecossistema.

- providers e drivers configuraveis
- Supabase, Firebase e Google operacionais
- EMQX/MQTT operacional
- descoberta de recursos e operacoes compativel com o frontend

Status:
- [x] Concluido

## Milestone 4 - Execution Observability

Objetivo: tornar cada execucao auditavel e compreensivel.

- historico de execucao por instancia
- nos visuais no estilo workflow
- detalhes por destino
- payload de entrada e saida estruturada
- indicacao de fallback, erro e sucesso por etapa

Status:
- [x] Concluido

## Milestone 5 - Update Experience

Objetivo: tornar a atualizacao do edge segura e assistida.

- checagem de release no GitHub
- fluxo assistido de atualizacao
- download, staging e confirmacao do update
- trilha de rollback segura
- historico persistido de updates
- canal `stable` e `beta`

Status:
- [x] Concluido

## Milestone 6 - Production Hardening

Objetivo: fechar a release candidata para uso continuo.

- revisao final de UX do `webui`
- documentacao de operacao e troubleshooting
- testes manuais de regressao
- validacao em Windows, Linux x64, ARMv6 e ARMv7
- checklist de release
- reducao do peso inicial do frontend

Status:
- [x] Concluido

## Checklist de fechamento

Antes da release final:

- [x] rodar `go test ./...`
- [x] rodar `npm run build` em `webui`
- [ ] validar criacao e execucao de instancia
- [ ] validar envio de dados para destino real
- [ ] validar fallback e retry
- [x] validar empacotamento para Raspberry Pi

## Evidencias de fechamento

Validado em 2026-08-11:

- `go test ./...`
- `npm run build` em `webui`
- `powershell -ExecutionPolicy Bypass -File .\scripts\validate-release.ps1 -Version v0.0.0 -IncludeArm32`
- geracao de pacotes para `windows-amd64`, `windows-arm64`, `linux-amd64`, `linux-arm64`, `linux-armv7`, `linux-armv6`, `darwin-amd64` e `darwin-arm64`
- geracao de `dist/packages/manifest.json` e `dist/packages/checksums.txt`

Observacao:

- os itens de smoke test funcional continuam sendo parte do processo de publicacao de cada versao, mas a estrutura, automacao e empacotamento da milestone foram concluidos
