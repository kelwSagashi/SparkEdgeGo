# Release Checklist

Use este checklist antes de publicar uma nova versao do SparkEdge Go.

## Baseline validada

Em 2026-08-11 foi validado o baseline automatizado com:

- `go test ./...`
- `npm run build` em `webui`
- `powershell -ExecutionPolicy Bypass -File .\scripts\validate-release.ps1 -Version v0.0.0 -IncludeArm32`
- geracao de `manifest.json` e `checksums.txt`

Este checklist continua valendo como smoke test por release. Os itens abaixo devem ser revisitados a cada nova versao publicada.

## 1. Validacao do backend

- [ ] rodar `go test ./...`
- [ ] validar subida local da API
- [ ] validar autenticacao JWT
- [ ] validar leitura de `config.yml`
- [ ] validar abertura da `webui` servida pelo binario Go

## 2. Validacao funcional

- [ ] criar servidor
- [ ] criar credencial
- [ ] criar script
- [ ] executar playground do script
- [ ] criar instancia
- [ ] validar destinos e mapeamentos
- [ ] validar envio para provider real
- [ ] validar fallback local
- [ ] validar retry do fallback

## 3. Validacao do updater

- [ ] verificar `GET /api/update/check`
- [ ] validar tela `/settings/update`
- [ ] validar download assistido
- [ ] validar apply assistido
- [ ] validar rollback assistido
- [ ] validar restart assistido ou plano de restart
- [ ] conferir historico persistido de updates

## 4. Validacao do frontend

- [ ] rodar `npm run build`
- [ ] abrir rotas principais
- [ ] validar `instances`
- [ ] validar `servers`
- [ ] validar `script-hub`
- [ ] validar `settings`
- [ ] validar `cloud`
- [ ] validar `advanced`
- [ ] validar `update`

## 5. Empacotamento

- [ ] gerar pacote Windows amd64
- [ ] gerar pacote Windows arm64
- [ ] gerar pacote Linux amd64
- [ ] gerar pacote Linux arm64
- [ ] gerar pacote Linux armv7
- [ ] gerar pacote Linux armv6
- [ ] gerar pacote macOS amd64
- [ ] gerar pacote macOS arm64
- [ ] validar `manifest.json`
- [ ] validar `checksums.txt`

## 6. Publicacao

- [ ] confirmar versao `vX.Y.Z`
- [ ] confirmar changelog
- [ ] conferir workflow de release no GitHub
- [ ] conferir assets publicados
- [ ] conferir release notes
