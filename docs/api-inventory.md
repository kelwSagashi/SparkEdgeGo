# Inventario da API TypeScript

Este documento mapeia a superficie HTTP atual do SparkEdge TypeScript para orientar a reescrita em Go.

No backend atual, toda rota `@RestController('/x')` entra sob `/api/x`, exceto controllers explicitamente registrados como root-level. Na base atual analisada, os controllers do CLI usam `@RestController`.

## Regras globais atuais

Rotas abertas pelo middleware global:

- `/api/auth`
- `/api/health`
- `/api/nodes`
- `/api/webhook`

Rotas protegidas pelo middleware global:

- `/api/instances`
- `/api/scripts`
- `/api/devices`
- `/api/servers`
- `/api/users`
- `/api/server-types`
- `/api/credentials`
- `/api/tags`
- `/api/executions`
- `/api/fallback`
- `/api/projects`

Observacoes importantes:

- `/api/cli`, `/api/adapters` e `/api/spark-cloud` nao aparecem nos prefixos protegidos atuais.
- O middleware de autenticacao tenta resolver usuario por cookie `spark_edge_token`, header `Authorization: Bearer ...` ou header `x-api-key`. Isso e comportamento esperado: o SparkEdge foi desenhado para aceitar diferentes formas de conexao, principalmente externas.
- Controllers podem escrever diretamente no `Response`, entao nem toda rota passa integralmente pelo wrapper JSON padrao.
- Existe um controller avancado de instancias em `instance-advanced.controller.ts`, mas ele nao e registrado no servidor atual. Na versao Go, ele deve ser tratado separadamente como `/api/instance-advanced` para evitar sobreposicao.

## Prioridades

- `P0`: essencial para a aplicacao operar.
- `P1`: importante para completar fluxos de uso.
- `P2`: suporte, diagnostico, onboarding ou refinamento.
- `Review`: precisa de decisao antes de portar, geralmente por sobreposicao ou comportamento incompleto.

## Health

| Metodo | Rota | Origem | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- | --- |
| GET | `/api/health` | `server.ts` | aberta | Migrada | Health check basico do servidor. |

## Auth

Origem: `packages/cli/src/auth/auth.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/register` | aberta | Migrada | Cria usuario local. |
| POST | `/api/auth/login` | aberta | Migrada | Cria cookie `spark_edge_token` e publica contexto via MQTT. |
| POST | `/api/auth/logout` | aberta | Migrada | Limpa cookie. |
| GET | `/api/auth/me` | aberta | Migrada | Retorna usuario resolvido pelo middleware, se houver. |
| POST | `/api/auth/generate-new-api-key/:userId` | aberta | Migrada | Gera nova API key. Rever se deve permanecer aberta. |

## Users

Origem: `packages/cli/src/users/user.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/users` | protegida | Migrada |
| GET | `/api/users/:id` | protegida | Migrada |
| GET | `/api/users/project/:id/:project` | protegida | Migrada |
| POST | `/api/users` | protegida | Migrada |
| PUT | `/api/users/:id` | protegida | Migrada |
| DELETE | `/api/users/:id` | protegida | Migrada |
| GET | `/api/users/:id/api-key` | protegida | Migrada |

## Projects

Origem: `packages/cli/src/projects/projects.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/projects` | protegida | Migrada |
| GET | `/api/projects/:id` | protegida | Migrada |
| POST | `/api/projects` | protegida | Migrada |
| PUT | `/api/projects/:id` | protegida | Migrada |
| DELETE | `/api/projects/:id` | protegida | Migrada |
| GET | `/api/projects/:id/members` | protegida | Migrada |
| POST | `/api/projects/:id/members` | protegida | Migrada |

## Devices

Origem: `packages/cli/src/devices/device.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/devices` | protegida | Migrada | Migrada em Go. |
| GET | `/api/devices/:id` | protegida | Migrada | Migrada em Go. |
| POST | `/api/devices` | protegida | Migrada | Migrada em Go com upsert. |
| PUT | `/api/devices/:id` | protegida | Migrada | Migrada em Go com upsert. |
| DELETE | `/api/devices/:id` | protegida | Migrada | Migrada em Go. |

## Scripts

Origem: `packages/cli/src/scripts/script.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/scripts` | protegida | Migrada | Lista scripts instalados. |
| GET | `/api/scripts/:id` | protegida | Migrada | Detalhe do script. |
| GET | `/api/scripts/:id/contents/:filename` | protegida | Migrada | Escreve diretamente no `Response`. |
| POST | `/api/scripts` | protegida | Migrada | Cria registro manual. |
| PUT | `/api/scripts/:id` | protegida | Migrada | Atualiza registro. |
| DELETE | `/api/scripts/:id` | protegida | Migrada | Remove registro. |
| POST | `/api/scripts/upload/inspect` | protegida | Migrada | Migrada em Go com multipart, zip e validacao de `sparkit`. |
| POST | `/api/scripts/upload/finalize` | protegida | Migrada | Migrada em Go: cria venv, instala requirements e captura schema via `--schema`. |
| GET | `/api/scripts/samples/list` | protegida | Migrada | Migrada em Go com `SPARKEDGE_SAMPLES_DIR` e fallback para `extensions/samples`. |
| GET | `/api/scripts/samples/:name/schema` | protegida | Migrada | Migrada em Go via Sparkit `--schema`. |
| POST | `/api/scripts/playground/run` | protegida | Migrada | Migrada em Go para `script_id` e `sample_name`. |

## Instances

Origem:

- `packages/cli/src/instances/instance.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/instances` | protegida | Migrada | Migrada em Go. |
| GET | `/api/instances/active` | protegida | Migrada | Migrada em Go. |
| GET | `/api/instances/project/:project_id` | protegida | Migrada | Migrada em Go. |
| GET | `/api/instances/:id` | protegida | Migrada | Migrada em Go com retorno de destinos e mappings. |
| POST | `/api/instances` | protegida | Migrada | Migrada em Go com normalizacao de payload, tags, destinos e mappings. |
| PUT | `/api/instances/:id` | protegida | Migrada | Migrada em Go com update parcial ou upsert completo com destinos. |
| DELETE | `/api/instances/:id` | protegida | Migrada | Migrada em Go. |
| POST | `/api/instances/:id/trigger` | protegida | Migrada | Migrada em Go com execucao Sparkit, registro em `instance_executions`, envio para providers e fallback local. |
| GET | `/api/instances/:id/executions` | protegida | Migrada | Migrada em Go usando `instance_executions`. |

## Instance Advanced

Origem:

- `packages/cli/src/instances/instance-advanced.controller.ts`

Observacao:

- Este controller existe no codigo, mas nao esta registrado no bootstrap do servidor atual.
- Na versao Go, ele deve ser exposto como `/api/instance-advanced` caso seus fluxos sejam preservados.

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/instance-advanced` | protegida | Migrada |
| GET | `/api/instance-advanced/project/:projectId` | protegida | Migrada |
| GET | `/api/instance-advanced/:id` | protegida | Migrada |
| POST | `/api/instance-advanced` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id` | protegida | Migrada |
| DELETE | `/api/instance-advanced/:id` | protegida | Migrada |
| POST | `/api/instance-advanced/:id/trigger` | protegida | Migrada |
| GET | `/api/instance-advanced/:id/executions` | protegida | Migrada |
| GET | `/api/instance-advanced/:id/executions/:executionId` | protegida | Migrada |
| GET | `/api/instance-advanced/:id/destinations` | protegida | Migrada |
| POST | `/api/instance-advanced/:id/destinations` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/destinations/:destinationId` | protegida | Migrada |
| DELETE | `/api/instance-advanced/:id/destinations/:destinationId` | protegida | Migrada |
| GET | `/api/instance-advanced/:id/available-fields` | protegida | Migrada |
| POST | `/api/instance-advanced/:id/mappings/test` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/destinations/:destinationId/mapping` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/active` | protegida | Migrada |
| GET | `/api/instance-advanced/:id/status` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/trigger-config` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/script-params` | protegida | Migrada |
| PUT | `/api/instance-advanced/:id/fallback-config` | protegida | Migrada |

## Executions

Origem: `packages/cli/src/executions/executions.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/executions` | protegida | Migrada | Migrada em Go. |
| GET | `/api/executions/:id` | protegida | Migrada | Migrada em Go. |
| GET | `/api/executions/instance/:instanceId` | protegida | Migrada | Migrada em Go como `/api/executions/instance/{instance_id}`. |

## Fallback

Origem: `packages/cli/src/fallback/fallback.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/fallback` | protegida | Migrada | Migrada em Go. |
| GET | `/api/fallback/stats` | protegida | Migrada | Migrada em Go. |
| POST | `/api/fallback/flush` | protegida | Migrada | Migrada em Go. |
| POST | `/api/fallback/:id/retry` | protegida | Migrada | Migrada em Go. |
| DELETE | `/api/fallback/:id` | protegida | Migrada | Migrada em Go. |

## Servers

Origem: `packages/cli/src/servers/server.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/servers` | protegida | Migrada |
| GET | `/api/servers/:id` | protegida | Migrada |
| GET | `/api/servers/:id/resources` | protegida | Migrada |
| GET | `/api/servers/:id/endpoints` | protegida | Migrada |
| POST | `/api/servers/execute` | protegida | Migrada |
| POST | `/api/servers` | protegida | Migrada |
| POST | `/api/servers/register` | protegida | Migrada |
| PUT | `/api/servers/:id` | protegida | Migrada |
| DELETE | `/api/servers/:id` | protegida | Migrada |

## Server Types

Origem: `packages/cli/src/servers/server-types.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/server-types` | protegida | Migrada |
| GET | `/api/server-types/:id` | protegida | Migrada |
| POST | `/api/server-types` | protegida | Migrada |
| PUT | `/api/server-types/:id` | protegida | Migrada |
| DELETE | `/api/server-types/:id` | protegida | Migrada |

## Credentials

Origem: `packages/cli/src/credentials/credentials.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/credentials/config/meta` | protegida | Migrada |
| GET | `/api/credentials` | protegida | Migrada |
| GET | `/api/credentials/:id` | protegida | Migrada |
| POST | `/api/credentials` | protegida | Migrada |
| PUT | `/api/credentials/:id` | protegida | Migrada |
| DELETE | `/api/credentials/:id` | protegida | Migrada |
| POST | `/api/credentials/test` | protegida | Migrada |

## Tags

Origem: `packages/cli/src/tags/tags.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/tags` | protegida | Migrada | Migrada em Go. |
| GET | `/api/tags/search` | protegida | Migrada | Migrada em Go. |
| POST | `/api/tags` | protegida | Migrada | Migrada em Go com upsert por nome/projeto. |
| DELETE | `/api/tags/:id` | protegida | Migrada | Migrada em Go. |

## Adapters

Origem: `packages/cli/src/instances/adapters.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/adapters/metadata` | nao protegida pelo middleware atual | Migrada | Migrada em Go via catalogo `server_types/auth_type`. |
| POST | `/api/adapters/:id/discover` | nao protegida pelo middleware atual | Migrada | Migrada em Go chamando `Adapter.Discover`. |

## Webhook

Origem: `packages/cli/src/webhook/webhook.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/webhook/:instanceId` | aberta | Migrada | Dispara instancia por webhook. |

## MQTT / CLI HTTP

Origem: `packages/cli/src/integrations/mqtt/cli.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/cli/onboarding` | nao protegida pelo middleware atual | Migrada | Dados de onboarding. |
| POST | `/api/cli/onboarding` | nao protegida pelo middleware atual | Migrada | Atualiza onboarding. |
| GET | `/api/cli/status` | nao protegida pelo middleware atual | Migrada | Status da conexao/cloud/MQTT. |
| POST | `/api/cli/pair` | nao protegida pelo middleware atual | Migrada | Pareamento. |
| POST | `/api/cli/connect` | nao protegida pelo middleware atual | Migrada | Login/conexao com cloud e EMQX. |
| POST | `/api/cli/disconnect` | nao protegida pelo middleware atual | Migrada | Desconecta preservando identidade. |
| POST | `/api/cli/reconnect` | nao protegida pelo middleware atual | Migrada | Reconecta com credenciais atuais. |
| POST | `/api/cli/remove` | nao protegida pelo middleware atual | Migrada | Remove vinculo cloud. |
| GET | `/api/cli/config` | nao protegida pelo middleware atual | Migrada | Migrada em Go lendo env/defaults. |
| PUT | `/api/cli/config` | nao protegida pelo middleware atual | Migrada | Migrada em Go como validacao; aplicacao efetiva via env/restart. |

## Spark Cloud Proxy

Origem: `packages/cli/src/integrations/spark-cloud/spark-cloud.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/spark-cloud/auth/login` | nao protegida pelo middleware atual | Migrada | Simulador local do Spark Cloud para login de provisionamento. |
| POST | `/api/spark-cloud/edges/register` | nao protegida pelo middleware atual | Migrada | Simulador local de registro de edge com credenciais MQTT/EMQX. |

## Blocos migrados

1. `health`, resposta padrao e middlewares.
2. `auth`, `users` e modelo de usuario local com JWT, cookie e API key.
3. `projects`, `devices`, `tags`, `scripts` e `credentials`.
4. `instances`, `instance-advanced`, scheduler, webhook e motor Sparkit.
5. `servers`, `server-types`, `adapters`, providers e drivers externos.
6. `executions` e fallback local com retry.
7. `cli` HTTP/MQTT, EMQX, provisionamento e simulador Spark Cloud.

## Checklist de decisoes

- Confirmar quais rotas atualmente abertas devem permanecer abertas em Go.
- Manter `/api/instances` e `/api/instance-advanced` separados na versao Go; o controller avancado foi consolidado em `/api/instance-advanced`.
- Definir se `/api/adapters` deve exigir usuario autenticado.
- Definir se `/api/cli/config` pode ser acessado sem auth.
- Padronizar handlers que escrevem diretamente no `Response`.
- Garantir que upload/finalize de scripts preserve a exigencia de `sparkit`.
