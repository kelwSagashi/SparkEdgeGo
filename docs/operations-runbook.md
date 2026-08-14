# Operational Runbook

Este runbook fecha a operacao padrao do SparkEdge Go para desenvolvimento, empacotamento, publicacao, instalacao em campo e atualizacao assistida.

## 1. Estrutura operacional esperada

Pacote extraido:

```text
sparkEdge/
  sparkedge.exe ou sparkedge
  sparkedge.db
  config.yml
  README.md
  version.txt
  webui/
    dist/
  config/
    .env.example
  updates/
    downloads/
    staging/
    backups/
```

Regras praticas:

- o binario sobe a API e serve a `webui/dist`;
- o SQLite principal continua local como `sparkedge.db`;
- a configuracao principal fica em `config.yml`;
- a pasta `updates/` e criada sob demanda pelo fluxo de update assistido.

## 2. Fluxo operacional de desenvolvimento

### Backend

```powershell
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
& 'C:\Program Files\Go\bin\go.exe' run ./cmd/sparkedge-api
```

### WebUI

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo\webui"
npm.cmd run dev
```

Padrao de uso:

- backend em `http://localhost:3009`;
- Vite em `http://localhost:5173`;
- chamadas `/api` indo para o backend local.

## 3. Fluxo operacional de release

### Validacao tecnica local

```powershell
./scripts/validate-release.ps1 -Version v0.0.0 -IncludeArm32
```

Esse comando valida:

- suite Go;
- build da WebUI;
- empacotamento multi-plataforma;
- `manifest.json`;
- `checksums.txt`.

### Empacotamento manual

```powershell
./scripts/build-release.ps1 -TargetOS windows -TargetArch amd64 -Version v0.0.0
```

### Publicacao automatica

O workflow de release atual:

- roda automaticamente em `push` para `main`;
- tambem pode ser acionado via `workflow_dispatch`;
- calcula a proxima versao semantica no formato `vMAJOR.MINOR.PATCH`;
- gera os pacotes:
  - `windows-amd64`
  - `windows-arm64`
  - `linux-amd64`
  - `linux-arm64`
  - `linux-armv7`
  - `linux-armv6`
  - `darwin-amd64`
  - `darwin-arm64`
- publica release no GitHub com:
  - arquivos `.zip`
  - `manifest.json`
  - `checksums.txt`

## 4. Smoke test da release

Apos extrair um pacote:

1. iniciar o binario dentro da pasta `sparkEdge`;
2. abrir `http://localhost:3009/api/health`;
3. abrir `http://localhost:3009/`;
4. confirmar que a WebUI abriu a partir de `webui/dist`;
5. validar que `config.yml` foi lido;
6. criar ou abrir uma entidade simples no sistema;
7. validar que o banco `sparkedge.db` foi criado ou reutilizado.

## 5. Instalacao em Raspberry Pi 3 armv7

Fluxo recomendado:

1. baixar o asset `linux-armv7`;
2. copiar o `.zip` para o Raspberry;
3. extrair o pacote;
4. entrar na pasta `sparkEdge`;
5. garantir permissao de execucao;
6. iniciar o binario.

Comandos tipicos:

```bash
chmod +x sparkedge
./sparkedge
```

Se quiser rodar em background:

```bash
nohup ./sparkedge > sparkedge.log 2>&1 &
```

Para descobrir o PID:

```bash
pgrep -af sparkedge
```

Para acompanhar o log:

```bash
tail -f sparkedge.log
```

Para parar:

```bash
pkill -f sparkedge
```

Ou, de forma mais controlada:

```bash
ps aux | grep sparkedge
kill <PID>
```

Se o processo nao encerrar:

```bash
kill -9 <PID>
```

Para producao em Raspberry, o caminho recomendado e usar `systemd` em vez de `nohup`, porque isso facilita:

- restart automatico;
- subida no boot;
- leitura de logs com `journalctl`;
- operacao padronizada com `systemctl start|stop|restart|status`.

## 6. Operacao offline e baixa conectividade

Na operacao de campo, considerar:

- o Edge deve continuar funcional sem depender do Cloud;
- falhas de envio para nuvem ou destinos externos nao devem impedir coleta local;
- fallback local e filas de sincronizacao devem ser observados pela WebUI;
- o update assistido deve ser usado em janelas controladas, nunca durante coleta critica.

Checklist minimo para ambiente remoto:

- `config.yml` revisado;
- conectividade MQTT e Cloud verificadas quando aplicavel;
- sincronizacao offline habilitada quando o cenario exigir;
- espaco em disco disponivel para `updates/`, `fallback` e banco local.

## 7. Update assistido em producao

Fluxo seguro:

1. abrir `/settings/update`;
2. executar verificacao de release;
3. baixar pacote compativel;
4. aplicar staging com backup;
5. revisar resultado;
6. reiniciar conforme plano sugerido;
7. se necessario, executar rollback.

Garantias esperadas:

- validacao por checksum quando disponivel;
- backup antes da troca;
- preservacao de `config.yml`;
- preservacao de `sparkedge.db`;
- historico persistido de updates.

## 8. Evidencias minimas de fechamento operacional

Para considerar uma versao operacionalmente fechada, manter:

- `go test ./...` passando;
- `npm run build` da WebUI passando;
- workflow de release gerando todos os targets;
- release com `manifest.json` e `checksums.txt`;
- smoke test local concluido;
- fluxo de update assistido validado ao menos em ambiente controlado.


## Atualizar manualmente no Linux

```bash
cd /home/rasplaber/sparkEdge
ZIP_PATH="$(find updates/downloads -type f -name '*.zip' | head -n 1)"
echo "$ZIP_PATH"
mkdir -p /tmp/sparkedge-manual-update
rm -rf /tmp/sparkedge-manual-update/*
unzip "$ZIP_PATH" -d /tmp/sparkedge-manual-update
pkill -f sparkedge
cd /home/rasplaber
cp -a sparkEdge sparkEdge.backup-$(date +%Y%m%d-%H%M%S)
cp -f /tmp/sparkedge-manual-update/sparkEdge/sparkedge /home/rasplaber/sparkEdge/
chmod +x /home/rasplaber/sparkEdge/sparkedge
rm -rf /home/rasplaber/sparkEdge/webui/dist
mkdir -p /home/rasplaber/sparkEdge/webui
cp -a /tmp/sparkedge-manual-update/sparkEdge/webui/dist /home/rasplaber/sparkEdge/webui/
cd /home/rasplaber/sparkEdge
nohup ./sparkedge > sparkedge.log 2>&1 &
```