# Mapa TypeScript para Go

Este documento existe para ajudar quem conhece a logica atual do SparkEdge em TypeScript, mas ainda esta se familiarizando com Go.

A ideia principal: em Go vamos preservar os mesmos conceitos do SparkEdge, mas com uma organizacao mais explicita. O Go nao usa decorators como `@Service()` e `@RestController()`, entao aquilo que hoje acontece por metadata e side effects passa a acontecer por composicao direta de pacotes, structs e interfaces.

## Visao rapida

| TypeScript atual | Go novo | Ideia equivalente |
| --- | --- | --- |
| `packages/cli/src/server.ts` | `internal/httpapi` + `cmd/sparkedge-api` | servidor HTTP e middlewares |
| `packages/cli/src/controller.registry.ts` | registro explicito de rotas por modulo | montagem modular das rotas |
| `packages/@spark-edge/di` | construtores e interfaces | injecao explicita de dependencias |
| `@Service()` | `NewService(...) *Service` | criacao de uma struct com dependencias |
| `@RestController('/x')` | `RegisterRoutes(mux, deps)` | modulo declara suas rotas |
| `@Get`, `@Post`, `@Put`, `@Delete` | `mux.HandleFunc("GET /api/x", handler)` | verbo e caminho HTTP |
| `response-helper.ts` | `internal/httpapi` response helpers | padrao de resposta e erro |
| `packages/cli/src/auth` | `internal/auth` + `internal/httpapi` | registro, login, JWT, cookie e API key |
| `packages/cli/src/users` | `internal/users` + `internal/httpapi` | usuarios, busca e API key |
| `packages/cli/src/projects` | `internal/projects` + `internal/httpapi` | projetos e membros |
| `packages/cli/src/devices` | `internal/devices` + `internal/httpapi` | dispositivos |
| `packages/cli/src/tags` | `internal/tags` + `internal/httpapi` | tags e vinculo com instancias |
| `packages/cli/src/instances/instance.controller.ts` | `internal/instances` + `internal/httpapi` | CRUD base de instancias |
| `packages/cli/src/executions/executions.controller.ts` | `internal/executions` + `internal/httpapi` | historico de execucoes |
| `packages/cli/src/scripts` | `internal/scripts` + `internal/httpapi` | scripts instalados e metadados |
| `packages/db` | `internal/sqlite` | banco local SQLite com ORM |
| Drizzle schemas | GORM models | definicao das tabelas por structs |
| `dbManager.repositories` | repositories Go usando GORM | acesso ao banco por entidade |
| `InstanceRunnerService` | `internal/runtime.Runner` | motor de execucao de instancias |
| `PythonVenvService` | `internal/python/sparkit.Executor` | execucao de scripts Python |
| `TemplateResolver` | `internal/runtime.TemplateResolver` | templates, contexto e JSONPath |
| `DestinationFactory` | `internal/providers.Registry` | cria adapter pelo tipo configurado |
| `BaseAdapter` | interface `providers.Adapter` | contrato para destinos externos |
| `ServerTypesTable` | `sqlite.serverTypeModel` | tipos de servidores externos |
| `AuthorizationsTypeTable` | `sqlite.authTypeModel` | metadados de autenticacao |
| `CredentialsTable` | `sqlite.credentialModel` | credenciais configuradas pelo usuario |
| `ServersTable` | `sqlite.serverModel` | servidores externos configurados |
| `ServerResourcesTable` | `sqlite.serverResourceModel` | recursos descobertos/configurados |
| `ResourceOperationsTable` | `sqlite.resourceOperationModel` | operacoes usadas por destinos |
| `spark-edge-core/modules/mqtt` | `internal/mqtt` | cliente EMQX, topicos e lifecycle |
| `CommandRegistry` | `cmd/sparkedge-cli` | comandos locais |

## Como pensar em Go

Em TypeScript, muita coisa do SparkEdge entra na aplicacao porque o arquivo e importado e os decorators registram metadata.

Em Go, vamos preferir algo mais direto:

```go
type ServerService struct {
	repo ServerRepository
}

func NewServerService(repo ServerRepository) *ServerService {
	return &ServerService{repo: repo}
}
```

Isso substitui o papel de `@Service()` e do container. A vantagem e que fica muito claro de onde cada dependencia vem.

## Controllers

No TypeScript:

```ts
@RestController("/servers")
export class ServersController {
  constructor(readonly serverService: ServerService) {}

  @Get("/")
  async list() {
    return this.serverService.list();
  }
}
```

No Go, a forma equivalente sera algo como:

```go
type ServerModule struct {
	service *ServerService
}

func (m *ServerModule) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/servers", m.list)
}

func (m *ServerModule) list(w http.ResponseWriter, r *http.Request) {
	// chama service, escreve resposta padronizada
}
```

O conceito e o mesmo: controller recebe service, rota chama metodo, resposta segue um padrao.

## Services

No TypeScript, services sao classes decoradas com `@Service()` e resolvidas pelo container.

No Go, services serao structs criadas por construtores:

```go
type InstanceService struct {
	instances InstanceRepository
	runner    *runtime.Runner
}

func NewInstanceService(instances InstanceRepository, runner *runtime.Runner) *InstanceService {
	return &InstanceService{instances: instances, runner: runner}
}
```

Isso deixa a dependencia visivel na assinatura.

## Repositories e SQLite

O banco principal continua sendo SQLite local.

No TypeScript:

- `packages/db/src/db/schemas.ts`
- `packages/db/src/repositories/*.repository.ts`
- `dbManager.instances.findById(...)`

No Go:

- `internal/sqlite`
- GORM como ORM;
- `github.com/glebarez/sqlite` como driver SQLite puro em Go;
- `AutoMigrate` para criar/atualizar as tabelas iniciais;
- repositories por entidade;
- metodos como `FindByID`, `Create`, `UpdateStatus`, `ListByInstance`.

O objetivo e que o codigo de dominio nao saiba detalhes de SQL nem de ORM. Ele fala com interfaces. A ORM fica confinada no pacote `internal/sqlite`.

Exemplo mental:

```go
type userModel struct {
	ID    string `gorm:"primaryKey"`
	Email string `gorm:"uniqueIndex;not null"`
}
```

Esse `userModel` e a descricao da tabela para a ORM. Ja `domain.User` continua sendo o objeto usado pela regra de negocio. Assim evitamos misturar detalhes de banco com dominio.

## Motor de execucao

Este e o coracao do SparkEdge.

No TypeScript, o fluxo esta em `InstanceRunnerService`:

1. busca instancia;
2. busca script;
3. resolve contexto e parametros;
4. executa Python;
5. captura `stdout` e `stderr`;
6. interpreta JSON;
7. aplica mapping;
8. envia para destinos;
9. usa fallback quando falha;
10. registra historico.

No Go, isso fica em `internal/runtime.Runner`.

O runner nao deve depender de HTTP. Ele deve ser uma peca reutilizavel pela API, CLI, webhook, scheduler e MQTT.

## Sparkit

Scripts Python devem usar Sparkit. Isso padroniza:

- `--schema`;
- `--input`;
- `--input-file`;
- stdout JSON;
- stderr estruturado;
- inputs e outputs declarados.

No Go, `internal/python/sparkit.Executor` sera a fronteira responsavel por chamar Python, passar entrada e normalizar saida.

## Scripts

No TypeScript, scripts ficam principalmente em:

- `packages/cli/src/scripts/script.controller.ts`;
- `packages/cli/src/scripts/script.service.ts`;
- `packages/db/src/repositories/downloadedScripts.repository.ts`;
- tabela `downloaded_scripts` em `schemas.ts`.

No Go, a primeira parte foi dividida assim:

- `internal/domain/script.go`: struct `DownloadedScript`;
- `internal/sqlite/models.go`: model GORM `downloadedScriptModel`;
- `internal/sqlite/scripts.go`: repository ORM da tabela `downloaded_scripts`;
- `internal/scripts/service.go`: regras de cadastro e remocao;
- `internal/httpapi/script_handlers.go`: rotas REST de CRUD.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `ScriptsController.list` | `handleScriptsList` | lista scripts instalados |
| `ScriptsController.getOne` | `handleScriptGet` | busca por `id` |
| `ScriptsController.create` | `scripts.Service.Create` + `handleScriptCreate` | cria registro manual |
| `ScriptsController.update` | `scripts.Service.Update` + `handleScriptUpdate` | atualiza metadados |
| `ScriptsController.remove` | `scripts.Service.Delete` + `handleScriptDelete` | remove registro e tenta remover pasta local |
| `ScriptsController.getFileContent` | `scripts.Service.FileContent` + `handleScriptFileContent` | le arquivo dentro da pasta do script |
| `ScriptsController.uploadInspect` | `scripts.Service.InspectZip` + `handleScriptUploadInspect` | extrai zip em temp e exige `sparkit` |
| `ScriptsController.uploadFinalize` | `scripts.Service.FinalizeUpload` + `handleScriptUploadFinalize` | move pasta, cria venv, instala requirements e captura schema |
| `ScriptsController.runPlayground` | `scripts.Service.RunPlayground` + `handleScriptPlaygroundRun` | executa script instalado com `--input-file` |
| `ScriptsController.listSamples` | `scripts.Service.ListSamples` + `handleScriptSamplesList` | lista samples locais |
| `ScriptsController.sampleSchema` | `scripts.Service.SampleSchema` + `handleScriptSampleSchema` | captura schema de sample |
| `dbManager.downloadedScripts` | `sqlite.ScriptsRepository` | persistencia local SQLite com GORM |

Campos JSON como `tags` e `schema_config` sao salvos por tipos auxiliares em `internal/sqlite/json.go`. Isso evita SQL manual e prepara a base para outros campos JSON de instancias, destinos, mappings e fallback.

Samples locais sao resolvidos pela variavel `SPARKEDGE_SAMPLES_DIR` ou por caminhos relativos como `extensions/samples`. O playground agora aceita tanto `script_id` quanto `sample_name`.

## Devices

No TypeScript, dispositivos ficam principalmente em:

- `packages/cli/src/devices/device.controller.ts`;
- `packages/cli/src/devices/device.service.ts`;
- `packages/db/src/repositories/devices.repository.ts`;
- tabela `devices` em `schemas.ts`.

No Go:

- `internal/domain/device.go`: struct `Device` e enums de conexao;
- `internal/sqlite/devices.go`: repository ORM da tabela `devices`;
- `internal/devices/service.go`: validacao e upsert;
- `internal/httpapi/device_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `DevicesController.list` | `handleDevicesList` | lista dispositivos |
| `DevicesController.getOne` | `handleDeviceGet` | busca por `id` |
| `DevicesController.create` | `devices.Service.Upsert` + `handleDeviceCreate` | cria ou atualiza conforme payload |
| `DevicesController.update` | `devices.Service.Upsert` + `handleDeviceUpdate` | usa `id` da URL |
| `DevicesController.remove` | `devices.Service.Delete` + `handleDeviceDelete` | remove dispositivo |
| `dbManager.devices` | `sqlite.DevicesRepository` | persistencia local SQLite com GORM |

O campo `others` continua sendo um JSON array, agora tipado como `[]DeviceOtherField` no dominio Go.

## Tags

No TypeScript, tags ficam principalmente em:

- `packages/cli/src/tags/tags.controller.ts`;
- `packages/cli/src/tags/tags.service.ts`;
- `packages/db/src/repositories/tags.repository.ts`;
- `packages/db/src/repositories/instanceTags.repository.ts`;
- tabelas `tags` e `instance_tags` em `schemas.ts`.

No Go:

- `internal/domain/tag.go`: structs `Tag` e `InstanceTag`;
- `internal/sqlite/tags.go`: repository ORM da tabela `tags`;
- `internal/sqlite/instance_tags.go`: repository ORM da tabela `instance_tags`;
- `internal/tags/service.go`: regras de tag e associacao com instancias;
- `internal/httpapi/tag_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `TagsController.list` | `handleTagsList` | lista tags, com `project_id` opcional |
| `TagsController.search` | `handleTagsSearch` | busca por `q` e `project_id` |
| `TagsController.create` | `tags.Service.Create` + `handleTagCreate` | upsert por nome/projeto |
| `TagsController.remove` | `tags.Service.Delete` + `handleTagDelete` | remove tag |
| `TagsService.findByInstance` | `tags.Service.FindByInstance` | usado internamente por instancias |
| `TagsService.linkTag` | `tags.Service.LinkTag` | usado internamente por instancias |
| `TagsService.unlinkTag` | `tags.Service.UnlinkTag` | usado internamente por instancias |
| `TagsService.syncTags` | `tags.Service.SyncTags` | usado internamente por instancias |
| `TagsService.findOrCreateByNames` | `tags.Service.FindOrCreateByNames` | cria tags ausentes por nome |

`instance_tags` ja existe no banco Go para preparar a migracao de instancias. Como a tabela `instances` ainda sera portada, mantivemos essa associacao pronta sem depender de foreign key rigida neste momento.

## Instances

No TypeScript, o controller basico de instancias fica em:

- `packages/cli/src/instances/instance.controller.ts`;
- `packages/cli/src/instances/instance.service.ts`;
- `packages/db/src/repositories/instances.repository.ts`;
- tabela `instances` em `schemas.ts`.

No Go, a primeira camada ficou assim:

- `internal/domain/instance.go`: struct `Instance`, status, trigger, fallback e erro;
- `internal/sqlite/instances.go`: repository ORM da tabela `instances`;
- `internal/instances/service.go`: normalizacao do payload antigo e sincronizacao de tags;
- `internal/httpapi/instance_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `InstanceController.list` | `handleInstancesList` | lista instancias |
| `InstanceController.listActive` | `handleInstancesActiveList` | lista ativas |
| `InstanceController.listByProject` | `handleInstancesByProjectList` | filtra por projeto |
| `InstanceController.getOne` | `handleInstanceGet` | retorna `{ instance, destinations }` com destinos e mapping |
| `InstanceController.create` | `instances.Service.Create` + `handleInstanceCreate` | cria instancia, sincroniza tags por nome e salva destinos |
| `InstanceController.update` | `instances.Service.Update` + `handleInstanceUpdate` | update parcial ou upsert completo quando houver destinos |
| `InstanceController.remove` | `instances.Service.Delete` + `handleInstanceDelete` | remove instancia |
| `InstanceController.triggerManual` | `handleInstanceTrigger` | executa Sparkit, aplica mappings e registra `instance_executions`; envio real/fallback ainda pendentes |
| `InstanceController.listExecutions` | `handleInstanceExecutionsList` | placeholder ate migrar executions |

O service Go ja preserva a normalizacao importante do TypeScript:

- aceita `project_id` e `projectId`;
- aceita `script_id` e `scriptId`;
- aceita `device_id` e `deviceId`;
- aceita `include_device_data` e `includeDeviceData`;
- mescla `script_inputs`/`scriptInputs` com `script_parameters`/`scriptParameters`;
- interpreta `fallback_config`/`fallbackConfig`;
- interpreta `error_config`/`errorConfig`;
- cria tags por nome e sincroniza `instance_tags`.
- salva `destinations` em `instance_destinations`;
- salva `mapping`, `data_mapping` ou `dataMapping` em `data_mappings`;
- aceita `retry_policy` e `retryPolicy` para politica de retry do destino.
- no runner, aplica `mapping`, `payload_template`/`payloadTemplate` e `custom_fields`/`customFields` antes do envio.

Ainda falta para instancias:

- `instance_executions`;
- integracao com fallback e providers.

## Providers e drivers

## Executions

No TypeScript, o historico de execucoes fica principalmente em:

- `packages/cli/src/executions/executions.controller.ts`;
- `packages/cli/src/instances/instance-execution.service.ts`;
- `packages/db/src/repositories/instanceExecutions.repository.ts`;
- tabela `instance_executions` em `schemas.ts`.

No Go:

- `internal/domain/execution.go`: struct `InstanceExecution` e `ExecutionLog`;
- `internal/sqlite/instance_executions.go`: repository ORM da tabela `instance_executions`;
- `internal/executions/service.go`: regras de listagem, busca, criacao e update de status;
- `internal/httpapi/execution_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `ExecutionsController.listAll` | `handleExecutionsList` | lista historico geral, com `limit` opcional |
| `ExecutionsController.getOne` | `handleExecutionGet` | busca execucao por `id` |
| `ExecutionsController.listByInstance` | `handleExecutionsByInstanceList` | lista historico por instancia |
| `dbManager.instanceExecutions.create` | `executions.Service.Create` | cria registro para o runner usar |
| `dbManager.instanceExecutions.updateStatus` | `executions.Service.UpdateStatus` | atualiza status, output, erro, logs e flags |

O endpoint `GET /api/instances/{id}/executions` agora tambem usa esse mesmo service, entao deixa de ser placeholder.

## Providers e drivers

No TypeScript, os adapters ficam em `packages/cli/src/instances/adapters` e drivers em `packages/cli/src/instances/drivers`.

No Go:

- `internal/providers.Registry` registra factories;
- cada provider implementa `providers.Adapter`;
- `Send`, `Test` e `Discover` preservam o contrato atual.
- `internal/sqlite/server_infrastructure.go` resolve `resource_operation_id` ate operation, resource, server e credential.
- `internal/providers/httpprovider` implementa os adapters HTTP equivalentes a `http-noauth`, `http-apikey`, `http-basicauth` e `http-bearer`.
- `internal/app.New` registra os adapters explicitamente, substituindo o papel do decorator `@CredentialAdapter()`.

Exemplo mental:

```go
type Adapter interface {
	Send(ctx context.Context, payload map[string]any) error
	Test(ctx context.Context, payload map[string]any) error
	Discover(ctx context.Context) ([]Resource, error)
}
```

Equivalencias praticas dos adapters HTTP:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `HttpDriver.request` | `httpprovider.Adapter.do` | monta URL, headers, body JSON e trata erro HTTP |
| `HttpNoAuthAdapter` | strategy `no_auth` | envia payload JSON sem credencial |
| `HttpApiKeyAdapter` | strategy `api_key` | aceita API key em header ou query |
| `HttpBasicAuthAdapter` | strategy `basic_auth` | monta header `Authorization: Basic ...` |
| `HttpBearerAdapter` | strategy `bearer_token` | monta header `Authorization: Bearer ...` |
| `AdapterRegistry.register` | `httpprovider.Register(registry)` | registro explicito no boot da aplicacao |
| `ServerTypeRegistry.syncWithDatabase` | `serverinfra.Service.SeedCatalog` | grava `server_types` no SQLite durante o boot |
| `AdapterRegistry.syncWithDatabase` | `serverinfra.Service.SeedCatalog` | grava `auth_type` no SQLite durante o boot |
| `CredentialsController.getMeta` | `GET /api/credentials/config/meta` | lista os tipos de autorizacao para o frontend atual |

No envio de destinos, o runner tenta selecionar o provider pela credencial (`auth_type_id`) quando ela existe. Isso preserva a ideia dos adapters TypeScript, onde o tipo de autorizacao escolhe a estrategia concreta. Quando nao ha credencial e o servidor e HTTP, o runner usa `no_auth`.

Equivalencias praticas do adapter Supabase:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `SupabaseServerType` | `supabaseprovider.ServerType` | cadastra o tipo `supabase` no catalogo |
| `SupabaseAdapter.metadata` | `supabaseprovider.AuthTypes` | cadastra campos `url` e `apiKey` |
| `SupabaseDriver.request` | `supabaseprovider.Adapter.request` | chama `/rest/v1/{table}` com `apikey` e bearer |
| `SupabaseAdapter.send` | `supabaseprovider.Adapter.Send` | suporta `insert`, `update`, `upsert` e `select` |
| `SupabaseAdapter.discover` | `supabaseprovider.Adapter.Discover` | le `definitions` do schema PostgREST |

Equivalencias praticas do adapter MongoDB:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `MongoServerType` | `mongoprovider.ServerType` | cadastra o tipo `mongodb` no catalogo |
| `MongoAdapter.metadata` | `mongoprovider.AuthTypes` | cadastra campos `uri` e `database` |
| `MongoDriver.connect` | `mongoprovider.connectMongo` | abre conexao com o driver oficial MongoDB para Go |
| `MongoAdapter.send` | `mongoprovider.Adapter.Send` | suporta `insertOne`, `updateOne`, `find` e `findOne` |
| `MongoAdapter.discover` | `mongoprovider.Adapter.Discover` | lista colecoes como recursos configuraveis |

## Middlewares

No Express, temos:

- CORS;
- cookie parser;
- JSON parser;
- autenticacao por cookie, bearer ou API key;
- protecao por prefixo de rota.

No Go, isso vira uma cadeia de middlewares:

```go
handler := recoverMiddleware(jsonMiddleware(authMiddleware(mux)))
```

O formato fica diferente, mas a ideia e a mesma: cada camada recebe a requisicao, faz seu trabalho e chama a proxima.

## Autenticacao

No TypeScript, a autenticacao fica em `packages/cli/src/auth` e conversa com o repositorio de usuarios em `packages/db`.

No Go, a migracao foi dividida em tres partes:

- `internal/auth`: regras de registro, login, senha, JWT e API key.
- `internal/sqlite`: persistencia local dos usuarios no SQLite.
- `internal/httpapi`: handlers REST e middleware que le cookie, bearer token ou `x-api-key`.

O ponto importante: o middleware pode encontrar uma credencial em varios lugares, como ja acontece no SparkEdge atual, mas rotas protegidas so seguem quando essa credencial e validada pelo servico de auth.

O token de sessao continua sendo JWT. Na versao Go, `internal/auth/token.go` usa JWT assinado com HS256, mantendo claims como `id`, `email`, `role` e expiracao.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `AuthService.register` | `auth.Service.Register` | cria usuario admin ativo no SQLite |
| `AuthService.login` | `auth.Service.Login` | valida senha e retorna usuario + JWT |
| `AuthService.verifyToken` | `auth.Service.VerifyToken` | valida JWT e carrega usuario |
| `AuthService.generateApiKey` | `auth.Service.GenerateAPIKey` | gera e salva API key do usuario |
| `findByApiKey` | `sqlite.UsersRepository.FindByAPIKey` | usado para conexoes externas |
| cookie `spark_edge_token` | cookie `spark_edge_token` | mantido para compatibilidade conceitual |
| `Authorization: Bearer` | `Authorization: Bearer` | mantido |
| `x-api-key` | `x-api-key` | mantido |

## Projetos

No TypeScript, projetos ficam principalmente em:

- `packages/cli/src/projects/projects.controller.ts`;
- `packages/db/src/repositories/projects.repository.ts`;
- `packages/db/src/repositories/projectMembers.repository.ts`;
- tabelas `projects` e `project_members` em `schemas.ts`.

No Go, dividimos assim:

- `internal/domain/project.go`: structs `Project` e `ProjectMember`;
- `internal/sqlite/models.go`: models GORM para criar as tabelas;
- `internal/sqlite/projects.go`: repository ORM da tabela `projects`;
- `internal/sqlite/project_members.go`: repository ORM da tabela `project_members`;
- `internal/projects/service.go`: regras de uso, como criar projeto e adicionar dono como membro;
- `internal/httpapi/project_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `ProjectsController.list` | `handleProjectsList` | lista projetos do usuario autenticado |
| `ProjectsController.getOne` | `handleProjectGet` | busca por `id` |
| `ProjectsController.create` | `projects.Service.Create` + `handleProjectCreate` | grava `owner_id` do usuario autenticado |
| `ProjectsController.update` | `projects.Service.Update` + `handleProjectUpdate` | atualiza campos editaveis |
| `ProjectsController.remove` | `projects.Service.Delete` + `handleProjectDelete` | remove projeto |
| `ProjectsController.listMembers` | `projects.Service.ListMembers` | lista membros |
| `ProjectsController.addMember` | `projects.Service.AddMember` | adiciona ou atualiza membro |
| `dbManager.projects` | `sqlite.ProjectsRepository` | persistencia local SQLite |
| `dbManager.projectMembers` | `sqlite.ProjectMembersRepository` | persistencia local SQLite |

Quando um projeto e criado pela API Go, o dono tambem e gravado em `project_members` com role `owner`. Isso deixa explicito no banco algo que o backend vai usar bastante quando migrarmos permissoes mais finas.

## Usuarios

No TypeScript, usuarios ficam principalmente em:

- `packages/cli/src/users/user.controller.ts`;
- `packages/cli/src/users/user.service.ts`;
- `packages/db/src/repositories/users.repository.ts`;
- tabela `users` em `schemas.ts`.

No Go, a divisao ficou assim:

- `internal/domain`: struct `User`;
- `internal/sqlite/users.go`: repository ORM da tabela `users`;
- `internal/users/service.go`: regras de usuarios;
- `internal/httpapi/user_handlers.go`: rotas REST.

Equivalencias praticas:

| TypeScript atual | Go novo | Observacao |
| --- | --- | --- |
| `UserController.list` | `handleUsersList` | lista usuarios |
| `UserController.getOne` | `handleUserGet` | busca por `id` |
| `UserController.getProject` | `handleUserProjectGet` | busca usuario + projeto por nome |
| `UserController.create` | `users.Service.Create` + `handleUserCreate` | cria usuario |
| `UserController.update` | `users.Service.Update` + `handleUserUpdate` | atualiza ou cria por `id`, preservando o comportamento de upsert |
| `UserController.remove` | `users.Service.Delete` + `handleUserDelete` | remove usuario |
| `UserController.createApiKey` | `users.Service.CreateAPIKey` + `handleUserCreateAPIKey` | gera nova API key |
| `dbManager.users` | `sqlite.UsersRepository` | persistencia local SQLite com GORM |

Os controllers Express de `credentials`, `servers`, `server-types`, `resources` e `operations` correspondem aos handlers em `internal/httpapi/server_infrastructure_handlers.go`. A camada equivalente aos services de infraestrutura esta em `internal/serverinfra/service.go`, que valida requests e delega persistencia aos repositorios GORM.

O cadastro local de servidor nao executa a integracao: `driver_key`, `credential_id` e a operacao resolvida sao a configuracao consumida futuramente pelo provider/driver.

No motor de instancias, `dispatchDestinations` corresponde ao envio pos-execucao do service TypeScript. Em Go, ele resolve `resource_operation_id` via `ResourceOperationsRepository`, seleciona o provider por `server.driver_key` e passa um `providers.Config` estruturado ao adapter.

Por seguranca, as respostas REST usam a funcao `publicUser`, que nao devolve `password_hash`. A rota de API key devolve apenas `{ userId, apiKey }`, como no comportamento antigo.

## CLI

No TypeScript, `CommandRegistry` decide entre `start`, `pair`, `status`, `connect`, `disconnect`, `remove`, `reconnect` e `provision`.

No Go, cada comando deve virar uma funcao ou subcomando em `cmd/sparkedge-cli`.

Primeiro mantemos simples:

- `sparkedge-cli start`
- `sparkedge-cli status`

Depois migramos os comandos de cloud/provisionamento.

## Regra de ouro da migracao

Sempre que voce olhar para uma classe TypeScript, procure responder:

- isso e entrada da aplicacao? Entao vai para `cmd`.
- isso e regra de negocio? Entao vai para `internal/domain` ou `internal/app`.
- isso fala com banco? Entao vai para `internal/sqlite`.
- isso fala com Python/Sparkit? Entao vai para `internal/python/sparkit`.
- isso executa instancia? Entao vai para `internal/runtime`.
- isso fala com servico externo configuravel? Entao vai para `internal/providers`.
- isso fala com EMQX? Entao vai para `internal/mqtt`.
- isso e HTTP? Entao vai para `internal/httpapi`.
