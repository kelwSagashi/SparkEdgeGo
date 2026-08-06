# Build e desenvolvimento

Este guia concentra os comandos para:

- rodar o SparkEdge Go em modo de desenvolvimento;
- gerar o executavel local;
- gerar executaveis para outros sistemas operacionais;
- gerar binarios para Raspberry Pi.

## Visao geral

Hoje o projeto tem dois binarios Go principais:

- `cmd/sparkedge-api`: servidor HTTP que atende a API e serve o frontend Vite compilado;
- `cmd/sparkedge-cli`: entrada de linha de comando para operacoes locais do Edge.

Importante: o frontend nao esta embutido no binario Go. Em producao, o executavel procura os arquivos compilados do frontend nestes lugares:

1. no caminho definido em `SPARKEDGE_FRONTEND_DIST`;
2. em `./frontend/dist` a partir do diretorio atual;
3. em `frontend/dist` ao lado do executavel;
4. em `../frontend/dist` relativo ao executavel.

Por isso, para distribuir a aplicacao com interface web, o binario e a pasta `frontend/dist` precisam ir juntos.

## Preparacao no Windows

Na maquina atual, se o cache padrao do Go der problema de permissao, use caches locais dentro do repositorio:

```powershell
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
```

Se o `go` nao estiver disponivel na sessao atual, use o executavel diretamente:

```powershell
& 'C:\Program Files\Go\bin\go.exe' version
```

## Modo dev

No modo dev, o mais confortavel e rodar backend e frontend separadamente.

### 1. Subir o backend Go

No diretorio raiz do repositorio:

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo"
$env:GOCACHE="$PWD\.gocache"
$env:GOMODCACHE="$PWD\.gomodcache"
go run ./cmd/sparkedge-api
```

Por padrao, a API sobe em `http://localhost:3009`.

Se quiser trocar a porta:

```powershell
$env:SPARKEDGE_HTTP_ADDR=":3010"
go run ./cmd/sparkedge-api
```

### 2. Subir o frontend Vite

Em outro terminal:

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo\frontend"
npm.cmd run dev
```

Por padrao, o frontend abre em `http://localhost:5173` e encaminha chamadas `/api` para `http://localhost:3009`.

Se quiser apontar o frontend para outra URL da API:

```powershell
$env:VITE_API_URL="http://localhost:3010"
npm.cmd run dev
```

### 3. Build do frontend para teste local de producao

Quando quiser testar o comportamento do binario Go servindo o frontend compilado:

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo\frontend"
npm.cmd run build
```

Depois volte para a raiz e rode o backend Go normalmente. Ele vai encontrar `frontend/dist` e servir a interface no mesmo host da API.

## Gerar executavel local

### Gerar o servidor principal

Na raiz do projeto:

```powershell
cd "C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo"
go build -o .\bin\sparkedge-api.exe .\cmd\sparkedge-api
```

### Gerar a CLI

```powershell
go build -o .\bin\sparkedge-cli.exe .\cmd\sparkedge-cli
```

### Gerar o frontend compilado

```powershell
cd .\frontend
npm.cmd run build
cd ..
```

### Estrutura recomendada para distribuicao local

```text
bin/
  sparkedge-api.exe
  sparkedge-cli.exe
frontend/
  dist/
    index.html
    assets/
```

### Rodar o executavel local

Se voce estiver na raiz do repositorio e o `frontend/dist` ja existir:

```powershell
.\bin\sparkedge-api.exe
```

Se quiser rodar o binario a partir de outro lugar, aponte explicitamente onde esta o frontend compilado:

```powershell
$env:SPARKEDGE_FRONTEND_DIST="C:\caminho\para\frontend\dist"
.\sparkedge-api.exe
```

## Cross-compilation para outros sistemas

Como o projeto usa `github.com/glebarez/sqlite`, o SQLite e implementado em Go puro, o que facilita bastante a geracao de binarios para outros alvos.

Regra geral:

- `GOOS` define o sistema operacional;
- `GOARCH` define a arquitetura;
- `GOARM` e usado em builds `linux/arm` de 32 bits;
- o frontend deve ser compilado uma vez e enviado junto com o binario.

Nos exemplos abaixo, o build e do servidor principal. Se quiser a CLI, basta trocar `.\cmd\sparkedge-api` por `.\cmd\sparkedge-cli`.

### Linux amd64

```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o .\dist\sparkedge-api-linux-amd64 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### Linux arm64

```powershell
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -o .\dist\sparkedge-api-linux-arm64 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### Windows amd64

```powershell
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o .\dist\sparkedge-api-windows-amd64.exe .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### Windows arm64

```powershell
$env:GOOS="windows"
$env:GOARCH="arm64"
go build -o .\dist\sparkedge-api-windows-arm64.exe .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### macOS Intel

```powershell
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -o .\dist\sparkedge-api-darwin-amd64 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### macOS Apple Silicon

```powershell
$env:GOOS="darwin"
$env:GOARCH="arm64"
go build -o .\dist\sparkedge-api-darwin-arm64 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

## Raspberry Pi

O Raspberry pode aparecer em mais de um formato. Os alvos mais comuns para o SparkEdge sao estes:

### Raspberry Pi 4 ou 5 com sistema 64 bits

Use `linux/arm64`:

```powershell
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -o .\dist\sparkedge-api-raspberry-pi-arm64 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### Raspberry Pi 3, Pi 4 ou Zero 2 W com sistema 32 bits

Use `linux/arm` com `GOARM=7`:

```powershell
$env:GOOS="linux"
$env:GOARCH="arm"
$env:GOARM="7"
go build -o .\dist\sparkedge-api-raspberry-pi-armv7 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
Remove-Item Env:GOARM
```

### Raspberry Pi Zero ou Raspberry Pi 1

Use `linux/arm` com `GOARM=6`:

```powershell
$env:GOOS="linux"
$env:GOARCH="arm"
$env:GOARM="6"
go build -o .\dist\sparkedge-api-raspberry-pi-armv6 .\cmd\sparkedge-api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
Remove-Item Env:GOARM
```

### Como saber qual build usar no Raspberry

No Raspberry de destino, estes comandos ajudam:

```bash
uname -m
getconf LONG_BIT
```

Leitura rapida:

- `aarch64` normalmente indica `linux/arm64`;
- `armv7l` normalmente indica `linux/arm` com `GOARM=7`;
- `armv6l` normalmente indica `linux/arm` com `GOARM=6`.

## Empacotamento para deploy

Para rodar em outro computador ou dispositivo, envie pelo menos:

1. o binario `sparkedge-api`;
2. a pasta `frontend/dist`;
3. os arquivos e diretorios de runtime que a aplicacao usar no ambiente;
4. a configuracao de ambiente necessaria, como `JWT_SECRET`, `SPARKEDGE_HTTP_ADDR`, `SPARKEDGE_FRONTEND_DIST`, `SPARKEDGE_SAMPLES_DIR` e integracoes externas.

Uma estrutura simples de deploy fica assim:

```text
sparkedge/
  sparkedge-api
  frontend/
    dist/
      index.html
      assets/
```

Se quiser, voce pode manter o frontend fora dessa estrutura e definir:

```bash
export SPARKEDGE_FRONTEND_DIST=/opt/sparkedge/frontend/dist
```

## Validacao rapida depois do build

Depois de gerar um binario de producao:

1. rode o executavel;
2. abra `/api/health`;
3. abra a raiz `/`;
4. confirme se o frontend carregou;
5. confirme se login e navegacao inicial estao funcionando.

## Observacoes

- O frontend em modo de producao depende de `frontend/dist`; sem isso, o servidor Go responde erro ao abrir a interface web.
- Para desenvolvimento do frontend, use `npm.cmd run dev`; para distribuicao, use `npm.cmd run build`.
- Se voce quiser automatizar isso depois, vale criar scripts como `build-local.ps1`, `build-linux-arm64.ps1` e `build-raspberry-armv7.ps1`.
