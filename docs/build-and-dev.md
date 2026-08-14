# Build e desenvolvimento

Este guia cobre:

- modo de desenvolvimento;
- build local;
- empacotamento de producao;
- cross-compilation para Windows, Linux, macOS e Raspberry Pi;
- automacao de release no GitHub.

## Estrutura de producao

A distribuicao de producao agora segue a ideia de pacote em pasta, com o binario e a interface lado a lado:

```text
sparkEdge/
  sparkedge.exe ou sparkedge
  sparkedge.db
  webui/
    dist/
      index.html
      assets/
  config/
    .env.example
  README.md
  version.txt
```

O banco SQLite permanece no comportamento atual da aplicacao: `sparkedge.db` no local padrao de execucao.

## Modo dev

### Backend Go

Na raiz do repositorio:

```powershell
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
& 'C:\Program Files\Go\bin\go.exe' run ./cmd/sparkedge-api
```

Por padrao a API sobe em `http://localhost:3009`.

### WebUI Vite

Em outro terminal:

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo\webui"
npm.cmd run dev
```

Por padrao a WebUI sobe em `http://localhost:5173` e encaminha `/api` para `http://localhost:3009`.

## Build manual da WebUI

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo\webui"
npm.cmd run build
```

O resultado vai para `webui/dist`.

## Build local do binario

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo"
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
& 'C:\Program Files\Go\bin\go.exe' build -o .\bin\sparkedge.exe .\cmd\sparkedge-api
```

## Como a aplicacao encontra a WebUI

Em runtime, o servidor procura os arquivos nesta ordem:

1. `SPARKEDGE_WEBUI_DIST`
2. `SPARKEDGE_FRONTEND_DIST` para compatibilidade temporaria
3. `./webui/dist`
4. `./frontend/dist` para compatibilidade temporaria
5. `webui/dist` ao lado do executavel

## Empacotamento de producao

O script de release prepara o pacote final com a estrutura pronta para distribuicao:

```powershell
./scripts/build-release.ps1 -TargetOS windows -TargetArch amd64 -Version v0.1.0
```

Exemplo de saida:

```text
dist/
  packages/
    sparkedge-v0.1.0-windows-amd64.zip
    staging/
      windows-amd64/
        sparkEdge/
          sparkedge.exe
          webui/
            dist/
          config/
            .env.example
          README.md
          version.txt
```

## Validacao automatizada de release

Antes de publicar uma release, voce pode rodar um fluxo tecnico automatizado:

```powershell
./scripts/validate-release.ps1 -Version v0.1.0
```

Esse fluxo:

- roda `go test ./...`
- roda `npm run build` em `webui`
- gera os pacotes principais de release
- gera `manifest.json` e `checksums.txt`

Para incluir tambem os alvos `armv7` e `armv6`:

```powershell
./scripts/validate-release.ps1 -Version v0.1.0 -IncludeArm32
```

Para usar somente como validacao tecnica sem empacotar:

```powershell
./scripts/validate-release.ps1 -Version v0.1.0 -SkipPackaging
```

## Alvos suportados

### Windows amd64

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-release.ps1 -TargetOS windows -TargetArch amd64 -Version v0.1.0
```

```powershell
./scripts/build-release.ps1 -TargetOS windows -TargetArch amd64 -Version dev
```

### Windows arm64

```powershell
./scripts/build-release.ps1 -TargetOS windows -TargetArch arm64 -Version dev
```

### Linux amd64

```powershell
./scripts/build-release.ps1 -TargetOS linux -TargetArch amd64 -Version dev
```

### Linux arm64

```powershell
./scripts/build-release.ps1 -TargetOS linux -TargetArch arm64 -Version dev
```

### macOS amd64

```powershell
./scripts/build-release.ps1 -TargetOS darwin -TargetArch amd64 -Version dev
```

### macOS arm64

```powershell
./scripts/build-release.ps1 -TargetOS darwin -TargetArch arm64 -Version dev
```

## Raspberry Pi

### Raspberry Pi 4 ou 5 64 bits

```powershell
./scripts/build-release.ps1 -TargetOS linux -TargetArch arm64 -Version dev
```

### Raspberry Pi 3, 4 ou Zero 2 W 32 bits

```powershell
./scripts/build-release.ps1 -TargetOS linux -TargetArch arm -GoArm 7 -Version dev
```

### Raspberry Pi Zero ou Pi 1

```powershell
./scripts/build-release.ps1 -TargetOS linux -TargetArch arm -GoArm 6 -Version dev
```

## Validacao rapida

Depois de gerar o pacote:

1. extraia o `.zip`;
2. rode o executavel dentro da pasta `sparkEdge`;
3. abra `/api/health`;
4. abra `/`;
5. confirme se a WebUI carregou a partir de `webui/dist`.

## GitHub Actions

O workflow [`release.yml`](../.github/workflows/release.yml) gera automaticamente pacotes versionados para:

- Windows amd64 e arm64;
- Linux amd64 e arm64;
- Linux armv7 e armv6;
- macOS amd64 e arm64.

Ele roda de duas formas:

1. automaticamente a cada `push` na branch `main`;
2. manualmente via `workflow_dispatch`, com override de versao e bump semantico.

Em cada execucao de release, os artefatos `.zip`, o `manifest.json` e o `checksums.txt` sao publicados na release do GitHub gerada automaticamente.

## Referencia operacional

Para o fluxo completo de operacao, instalacao em campo, smoke test e update assistido, veja [docs/operations-runbook.md](./operations-runbook.md).
