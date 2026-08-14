# Troubleshooting

Este guia resume os problemas mais comuns no SparkEdge Go e como investigar cada um.

## API nao sobe

Verifique:

- `config.yml`
- porta configurada em `server.port`
- logs do processo
- se outro processo esta usando a porta

Teste rapido:

```powershell
netstat -ano | findstr 3009
```

## WebUI nao abre

Verifique:

- se `webui/dist` existe
- se o binario esta ao lado da pasta `webui`
- se a API responde em `/api/health`

Teste rapido:

```powershell
curl http://localhost:3009/api/health
```

## Build do frontend falha

Verifique:

- `node_modules` instalado
- versao do Node.js
- erros de tipagem ou import

Teste rapido:

```powershell
cd webui
npm.cmd run build
```

## Script Python nao executa

Verifique:

- se o `venv` foi criado
- se `requirements.txt` foi instalado
- se o `sparkit` esta no ambiente
- se o script principal esta correto

Pontos de analise:

- saida do playground
- historico da execucao da instancia
- pasta local do script

## Destino nao recebe payload

Verifique:

- mapeamento da instancia
- operacao selecionada no destino
- credencial associada ao servidor
- payload resolvido no historico da execucao

Pontos de analise:

- `input_payload`
- `output_payload`
- `destination_details`
- logs tecnicos da execucao

## Update assistido nao encontra release

Verifique:

- `update.enabled`
- `update.repo`
- `update.channel`
- conectividade com GitHub
- existencia de asset compativel com a plataforma

Pontos de analise:

- `/api/update/check`
- `updates/state.json`
- `manifest.json` da release

## Update assistido falha no apply

Verifique:

- nome do pacote `.zip`
- estrutura esperada `sparkEdge/`
- existencia de `webui/dist`
- permissao para trocar arquivos

Pontos de analise:

- pasta `updates/staging`
- pasta `updates/backups`
- scripts de apply e rollback

## Raspberry Pi nao inicia

Verifique:

- arquitetura correta do pacote
- permissao de execucao do binario
- compatibilidade do sistema operacional
- se foi usado o asset `linux-armv7` para Raspberry Pi 3 32 bits
- se o pacote foi extraido preservando a pasta `webui/dist`

Teste rapido:

```bash
chmod +x sparkedge
./sparkedge
```

## Release automatica nao apareceu no GitHub

Verifique:

- se houve `push` real na branch `main`
- se o workflow `release.yml` executou
- se todos os 8 targets foram gerados
- se `manifest.json` e `checksums.txt` foram publicados

Pontos de analise:

- aba Actions do GitHub
- release criada com tag `vMAJOR.MINOR.PATCH`
- assets `linux-armv7` e `linux-armv6`

## Update assistido preservou config mas nao refletiu a mudanca esperada

Verifique:

- se o `config.yml` local foi preservado de proposito
- se a mudanca esperada estava no binario ou na `webui/dist`
- se o restart foi realmente executado apos o apply

Pontos de analise:

- `updates/backups`
- `updates/staging`
- tela `/settings/update`
