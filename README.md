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
- `internal/devices`: cadastro de dispositivos usados no contexto das instancias.
- `internal/tags`: tags de organizacao e vinculo com instancias.
- `internal/instances`: cadastro, configuracao, destinos e mappings de instancias.
- `internal/executions`: historico de execucao das instancias.
- `internal/runtime`: motor de execucao de instancias.
- `internal/python/sparkit`: integracao com scripts Sparkit.
- `internal/providers`: registry de destinos externos.
- `internal/mqtt`: fronteira EMQX.
- `webui`: frontend Vite migrado para dentro do repositorio Go.

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
- `GET /api/devices`
- `POST /api/devices`
- `GET /api/devices/{id}`
- `PUT /api/devices/{id}`
- `DELETE /api/devices/{id}`
- `GET /api/tags`
- `GET /api/tags/search`
- `POST /api/tags`
- `DELETE /api/tags/{id}`
- `GET /api/instances`
- `GET /api/instances/active`
- `GET /api/instances/project/{project_id}`
- `POST /api/instances`
- `GET /api/instances/{id}`
- `PUT /api/instances/{id}`
- `DELETE /api/instances/{id}`
- `POST /api/instances/{id}/trigger`
- `GET /api/instances/{id}/executions`
- `GET /api/executions`
- `GET /api/executions/{id}`
- `GET /api/executions/instance/{instance_id}`
- `GET /api/scripts`
- `POST /api/scripts`
- `GET /api/scripts/{id}`
- `PUT /api/scripts/{id}`
- `DELETE /api/scripts/{id}`
- `GET /api/scripts/{id}/contents/{filename}`
- `POST /api/scripts/upload/inspect`
- `POST /api/scripts/upload/finalize`
- `POST /api/scripts/playground/run`
- `GET /api/scripts/samples/list`
- `GET /api/scripts/samples/{name}/schema`
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

A infraestrutura de providers ja possui tabelas SQLite/GORM para `server_types`, `auth_type`, `credentials`, `servers`, `server_resources` e `resource_operations`. Isso prepara o caminho para o runner resolver `resource_operation_id` ate server/resource/credential antes de chamar adapters concretos.

O runner agora resolve cada `resource_operation_id` pelo SQLite antes do dispatch. O `driver_key` do servidor seleciona o provider e o adapter recebe configuracoes separadas de servidor, recurso, operacao e credencial.

A WebUI Vite ja foi trazida para `webui/`, ganhou tipos locais proprios e agora gera `webui/dist`, que e servida em disco no mesmo host do backend, inclusive nos pacotes de producao.

A arvore de rotas HTTP foi adaptada para conviver com o `ServeMux` moderno do Go: familias com paths dinamicos mais ambiguos, como `instances`, `instance-advanced` e `scripts`, passaram a usar dispatch por prefixo para evitar conflitos de bootstrap.

Os endpoints de credenciais e servidores estao organizados em `internal/serverinfra` e conectados ao `httpapi`. Eles representam o cadastro local das integracoes; a comunicacao com cada servico externo continuara sendo responsabilidade dos drivers/providers.

O CRUD de scripts ja grava `downloaded_scripts`, incluindo campos JSON como `tags` e `schema_config`. O fluxo de upload tambem ja valida `requirements.txt` com `sparkit`, cria venv, instala dependencias, captura schema via `--schema` e executa playground com `--input-file`.

Samples locais podem ser apontados pela variavel `SPARKEDGE_SAMPLES_DIR`. Se ela nao existir, a aplicacao tenta localizar `extensions/samples` em caminhos relativos comuns ao ambiente de desenvolvimento.

Instancias ja possuem CRUD, listagem por projeto, listagem de ativas, sincronizacao de tags por nome, persistencia de `instance_destinations`, persistencia de `data_mappings` e trigger manual com execucao Sparkit registrada em `instance_executions`. O runner ja aplica mapping/template/custom fields dos destinos, registra payloads mapeados, envia para providers reais e usa fallback local com retry quando necessario.

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

## Build e execucao

O guia detalhado para modo dev, geracao de executavel local e cross-compilation esta em [docs/build-and-dev.md](./docs/build-and-dev.md).

## Proximas ondas

1. Validar navegacao ponta a ponta do frontend servido pelo Go, cobrindo login, instancias, scripts, dados locais e telas de configuracao.
2. Consolidar a camada HTTP do frontend, reduzindo a duplicacao entre `server.service.ts` e `rest-api-client/*`.
3. Executar testes funcionais reais com scripts `sparkit`, EMQX/MQTT e providers externos com credenciais verdadeiras.
4. Revisar outras familias de rotas dinamicas para manter compatibilidade com o `ServeMux` do Go sem depender de comportamento ambiguo.
5. Empacotar distribuicao local com `webui/dist`, binario Go e configuracao de producao.
