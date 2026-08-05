# SparkEdge Go

Nova base do SparkEdge em Go.

Este repositorio nasce para substituir o backend TypeScript do SparkEdge, preservando o frontend Vite e mantendo os conceitos atuais do sistema:

- banco principal local em SQLite;
- manipulacao do SQLite por ORM, evitando SQL manual nos repositories;
- execucao de scripts Python padronizada por Sparkit;
- MQTT via EMQX;
- providers e drivers plugaveis para destinos externos;
- servidor REST modular;
- CLI local;
- motor de execucao de instancias como componente de primeira classe.

## Estado atual

Esta base ainda esta no inicio da migracao, mas ja possui a fundacao do repositorio Go e a primeira fatia funcional de backend:

- `cmd/sparkedge-api`: entrada do servidor REST.
- `cmd/sparkedge-cli`: entrada da CLI.
- `internal/app`: composicao da aplicacao.
- `internal/domain`: entidades e contratos de dominio.
- `internal/httpapi`: servidor HTTP e padrao de resposta.
- `internal/sqlite`: fronteira do banco local usando GORM sobre SQLite.
- `internal/auth`: registro, login, JWT, verificacao de token e API key.
- `internal/users`: regras de usuarios, perfil e API key.
- `internal/projects`: regras de projetos e membros.
- `internal/scripts`: cadastro local de scripts instalados.
- `internal/runtime`: motor de execucao de instancias.
- `internal/python/sparkit`: integracao com scripts Sparkit.
- `internal/providers`: registry de destinos externos.
- `internal/mqtt`: fronteira EMQX.

Rotas ja iniciadas:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`
- `GET /api/health`
- `GET /api/users`
- `POST /api/users`
- `GET /api/users/project/{id}/{project}`
- `GET /api/users/{id}`
- `PUT /api/users/{id}`
- `DELETE /api/users/{id}`
- `GET /api/users/{id}/api-key`
- `GET /api/scripts`
- `POST /api/scripts`
- `GET /api/scripts/{id}`
- `PUT /api/scripts/{id}`
- `DELETE /api/scripts/{id}`
- `GET /api/projects`
- `POST /api/projects`
- `GET /api/projects/{id}`
- `PUT /api/projects/{id}`
- `DELETE /api/projects/{id}`
- `GET /api/projects/{id}/members`
- `POST /api/projects/{id}/members`

O middleware de autenticacao preserva a ideia original do SparkEdge: tentar resolver identidade por cookie, header `Authorization: Bearer ...` ou `x-api-key`. Em rotas protegidas, a credencial precisa ser validada pelo servico de auth antes da requisicao seguir.

O fluxo de autenticacao tambem preserva o comportamento antigo de garantir um projeto local `PERSONAL` para o usuario.

O token de sessao permanece sendo JWT assinado com HS256. Ele e retornado no login e tambem gravado no cookie `spark_edge_token`.

Na camada de banco, os schemas iniciais ficam em structs Go usadas pela ORM. Isso deixa a migracao mais parecida com o Drizzle do projeto TypeScript: as tabelas sao descritas como tipos e os repositories manipulam entidades sem escrever SQL diretamente.

O CRUD de scripts ja grava `downloaded_scripts`, incluindo campos JSON como `tags` e `schema_config`. As rotas de upload, inspect, finalize e playground ainda serao migradas em uma onda propria, junto da integracao Sparkit/Python.

## Ambiente Go no Windows

Se `go` ainda nao estiver no PATH da sessao atual, use o executavel direto:

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./...
```

Em ambientes com permissao restrita para o cache padrao, aponte o cache para dentro do repositorio:

```powershell
$env:GOCACHE="$PWD\.gocache"
& 'C:\Program Files\Go\bin\go.exe' test ./...
```

## Proximas ondas

1. Inventariar rotas e contratos do backend TypeScript.
2. Migrar schema SQLite e repositorios.
3. Portar o motor de execucao de instancias.
4. Portar providers/drivers.
5. Portar API REST e CLI.
6. Reapontar o frontend Vite.
