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

Teste rapido:

```bash
chmod +x sparkedge
./sparkedge
```
