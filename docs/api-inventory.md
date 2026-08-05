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
| GET | `/api/health` | `server.ts` | aberta | P0 | Health check basico do servidor. |

## Auth

Origem: `packages/cli/src/auth/auth.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/register` | aberta | P0 | Cria usuario local. |
| POST | `/api/auth/login` | aberta | P0 | Cria cookie `spark_edge_token` e publica contexto via MQTT. |
| POST | `/api/auth/logout` | aberta | P0 | Limpa cookie e contexto MQTT. |
| GET | `/api/auth/me` | aberta | P0 | Retorna usuario resolvido pelo middleware, se houver. |
| POST | `/api/auth/generate-new-api-key/:userId` | aberta | P1 | Gera nova API key. Rever se deve permanecer aberta. |

## Users

Origem: `packages/cli/src/users/user.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/users` | protegida | P1 |
| GET | `/api/users/:id` | protegida | P1 |
| GET | `/api/users/project/:id/:project` | protegida | P1 |
| POST | `/api/users` | protegida | P1 |
| PUT | `/api/users/:id` | protegida | P1 |
| DELETE | `/api/users/:id` | protegida | P1 |
| GET | `/api/users/:id/api-key` | protegida | P1 |

## Projects

Origem: `packages/cli/src/projects/projects.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/projects` | protegida | P0 |
| GET | `/api/projects/:id` | protegida | P0 |
| POST | `/api/projects` | protegida | P0 |
| PUT | `/api/projects/:id` | protegida | P0 |
| DELETE | `/api/projects/:id` | protegida | P1 |
| GET | `/api/projects/:id/members` | protegida | P1 |
| POST | `/api/projects/:id/members` | protegida | P1 |

## Devices

Origem: `packages/cli/src/devices/device.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/devices` | protegida | P0 | Migrada em Go. |
| GET | `/api/devices/:id` | protegida | P0 | Migrada em Go. |
| POST | `/api/devices` | protegida | P0 | Migrada em Go com upsert. |
| PUT | `/api/devices/:id` | protegida | P0 | Migrada em Go com upsert. |
| DELETE | `/api/devices/:id` | protegida | P1 | Migrada em Go. |

## Scripts

Origem: `packages/cli/src/scripts/script.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/scripts` | protegida | P0 | Lista scripts instalados. |
| GET | `/api/scripts/:id` | protegida | P0 | Detalhe do script. |
| GET | `/api/scripts/:id/contents/:filename` | protegida | P1 | Escreve diretamente no `Response`. |
| POST | `/api/scripts` | protegida | P1 | Cria registro manual. |
| PUT | `/api/scripts/:id` | protegida | P1 | Atualiza registro. |
| DELETE | `/api/scripts/:id` | protegida | P1 | Remove registro. |
| POST | `/api/scripts/upload/inspect` | protegida | P0 | Migrada em Go com multipart, zip e validacao de `sparkit`. |
| POST | `/api/scripts/upload/finalize` | protegida | P0 | Migrada em Go: cria venv, instala requirements e captura schema via `--schema`. |
| GET | `/api/scripts/samples/list` | protegida | P2 | Migrada em Go com `SPARKEDGE_SAMPLES_DIR` e fallback para `extensions/samples`. |
| GET | `/api/scripts/samples/:name/schema` | protegida | P2 | Migrada em Go via Sparkit `--schema`. |
| POST | `/api/scripts/playground/run` | protegida | P1 | Migrada em Go para `script_id` e `sample_name`. |

## Instances

Origem:

- `packages/cli/src/instances/instance.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/instances` | protegida | P0 | Migrada em Go. |
| GET | `/api/instances/active` | protegida | P1 | Migrada em Go. |
| GET | `/api/instances/project/:project_id` | protegida | P0 | Migrada em Go. |
| GET | `/api/instances/:id` | protegida | P0 | Migrada em Go; destinos ainda vazios ate portar mappings/destinations. |
| POST | `/api/instances` | protegida | P0 | Migrada em Go com normalizacao de payload e tags. |
| PUT | `/api/instances/:id` | protegida | P0 | Migrada em Go com update parcial. |
| DELETE | `/api/instances/:id` | protegida | P0 | Migrada em Go. |
| POST | `/api/instances/:id/trigger` | protegida | P0 | Placeholder em Go ate ligar runner real. |
| GET | `/api/instances/:id/executions` | protegida | P0 | Placeholder em Go ate migrar executions. |

## Instance Advanced

Origem:

- `packages/cli/src/instances/instance-advanced.controller.ts`

Observacao:

- Este controller existe no codigo, mas nao esta registrado no bootstrap do servidor atual.
- Na versao Go, ele deve ser exposto como `/api/instance-advanced` caso seus fluxos sejam preservados.

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/instance-advanced` | protegida | Review |
| GET | `/api/instance-advanced/project/:projectId` | protegida | Review |
| GET | `/api/instance-advanced/:id` | protegida | Review |
| POST | `/api/instance-advanced` | protegida | Review |
| PUT | `/api/instance-advanced/:id` | protegida | Review |
| DELETE | `/api/instance-advanced/:id` | protegida | Review |
| POST | `/api/instance-advanced/:id/trigger` | protegida | Review |
| GET | `/api/instance-advanced/:id/executions` | protegida | Review |
| GET | `/api/instance-advanced/:id/executions/:executionId` | protegida | P1 |
| GET | `/api/instance-advanced/:id/destinations` | protegida | P0 |
| POST | `/api/instance-advanced/:id/destinations` | protegida | P0 |
| PUT | `/api/instance-advanced/:id/destinations/:destinationId` | protegida | P0 |
| DELETE | `/api/instance-advanced/:id/destinations/:destinationId` | protegida | P0 |
| GET | `/api/instance-advanced/:id/available-fields` | protegida | P1 |
| POST | `/api/instance-advanced/:id/mappings/test` | protegida | Review |
| PUT | `/api/instance-advanced/:id/destinations/:destinationId/mapping` | protegida | P0 |
| PUT | `/api/instance-advanced/:id/active` | protegida | P0 |
| GET | `/api/instance-advanced/:id/status` | protegida | P0 |
| PUT | `/api/instance-advanced/:id/trigger-config` | protegida | P0 |
| PUT | `/api/instance-advanced/:id/script-params` | protegida | P0 |
| PUT | `/api/instance-advanced/:id/fallback-config` | protegida | P0 |

## Executions

Origem: `packages/cli/src/executions/executions.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/executions` | protegida | P1 |
| GET | `/api/executions/:id` | protegida | P1 |
| GET | `/api/executions/instance/:instanceId` | protegida | P1 |

## Fallback

Origem: `packages/cli/src/fallback/fallback.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/fallback` | protegida | P0 |
| GET | `/api/fallback/stats` | protegida | P0 |
| POST | `/api/fallback/flush` | protegida | P0 |
| POST | `/api/fallback/:id/retry` | protegida | P0 |
| DELETE | `/api/fallback/:id` | protegida | P1 |

## Servers

Origem: `packages/cli/src/servers/server.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/servers` | protegida | P0 |
| GET | `/api/servers/:id` | protegida | P0 |
| GET | `/api/servers/:id/resources` | protegida | P0 |
| GET | `/api/servers/:id/endpoints` | protegida | P1 |
| POST | `/api/servers/execute` | protegida | P0 |
| POST | `/api/servers` | protegida | P0 |
| POST | `/api/servers/register` | protegida | P0 |
| PUT | `/api/servers/:id` | protegida | P0 |
| DELETE | `/api/servers/:id` | protegida | P1 |

## Server Types

Origem: `packages/cli/src/servers/server-types.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/server-types` | protegida | P0 |
| GET | `/api/server-types/:id` | protegida | P0 |
| POST | `/api/server-types` | protegida | P1 |
| PUT | `/api/server-types/:id` | protegida | P1 |
| DELETE | `/api/server-types/:id` | protegida | P1 |

## Credentials

Origem: `packages/cli/src/credentials/credentials.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/credentials/config/meta` | protegida | P0 |
| GET | `/api/credentials` | protegida | P0 |
| GET | `/api/credentials/:id` | protegida | P0 |
| POST | `/api/credentials` | protegida | P0 |
| PUT | `/api/credentials/:id` | protegida | P0 |
| DELETE | `/api/credentials/:id` | protegida | P1 |
| POST | `/api/credentials/test` | protegida | P0 |

## Tags

Origem: `packages/cli/src/tags/tags.controller.ts`

| Metodo | Rota | Auth | Prioridade |
| --- | --- | --- | --- |
| GET | `/api/tags` | protegida | P1 | Migrada em Go. |
| GET | `/api/tags/search` | protegida | P1 | Migrada em Go. |
| POST | `/api/tags` | protegida | P1 | Migrada em Go com upsert por nome/projeto. |
| DELETE | `/api/tags/:id` | protegida | P1 | Migrada em Go. |

## Adapters

Origem: `packages/cli/src/instances/adapters.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/adapters/metadata` | nao protegida pelo middleware atual | P0 | Lista metadata dos adapters registrados. |
| POST | `/api/adapters/:id/discover` | nao protegida pelo middleware atual | P0 | Descobre recursos de adapter. Rever auth no Go. |

## Webhook

Origem: `packages/cli/src/webhook/webhook.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/webhook/:instanceId` | aberta | P0 | Dispara instancia por webhook. |

## MQTT / CLI HTTP

Origem: `packages/cli/src/integrations/mqtt/cli.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| GET | `/api/cli/onboarding` | nao protegida pelo middleware atual | P2 | Dados de onboarding. |
| POST | `/api/cli/onboarding` | nao protegida pelo middleware atual | P2 | Atualiza onboarding. |
| GET | `/api/cli/status` | nao protegida pelo middleware atual | P0 | Status da conexao/cloud/MQTT. |
| POST | `/api/cli/pair` | nao protegida pelo middleware atual | P0 | Pareamento. |
| POST | `/api/cli/connect` | nao protegida pelo middleware atual | P0 | Login/conexao com cloud e EMQX. |
| POST | `/api/cli/disconnect` | nao protegida pelo middleware atual | P1 | Desconecta preservando identidade. |
| POST | `/api/cli/reconnect` | nao protegida pelo middleware atual | P1 | Reconecta com credenciais atuais. |
| POST | `/api/cli/remove` | nao protegida pelo middleware atual | P1 | Remove vinculo cloud. |
| GET | `/api/cli/config` | nao protegida pelo middleware atual | P1 | Le config local. |
| PUT | `/api/cli/config` | nao protegida pelo middleware atual | P1 | Atualiza config local. |

## Spark Cloud Proxy

Origem: `packages/cli/src/integrations/spark-cloud/spark-cloud.controller.ts`

| Metodo | Rota | Auth | Prioridade | Observacao |
| --- | --- | --- | --- | --- |
| POST | `/api/spark-cloud/auth/login` | nao protegida pelo middleware atual | P1 | Proxy/login no Spark Cloud. |
| POST | `/api/spark-cloud/edges/register` | nao protegida pelo middleware atual | P1 | Registro de edge. |

## Ordem sugerida para portar

1. `health`, resposta padrao e middlewares.
2. `auth` e modelo de usuario local.
3. `projects`, `devices`, `scripts` e `credentials`.
4. `instances` com motor de execucao real.
5. `instance-advanced` como modulo separado, caso ainda seja necessario para fluxos avancados.
6. `servers`, `server-types`, `adapters` e providers.
7. `executions` e `fallback`.
8. `cli` HTTP/MQTT e Spark Cloud.
9. `users`, `tags` e endpoints auxiliares.

## Checklist de decisoes

- Confirmar quais rotas atualmente abertas devem permanecer abertas em Go.
- Manter `/api/instances` e `/api/instance-advanced` separados na versao Go, caso o controller avancado seja portado.
- Definir se `/api/adapters` deve exigir usuario autenticado.
- Definir se `/api/cli/config` pode ser acessado sem auth.
- Padronizar handlers que escrevem diretamente no `Response`.
- Garantir que upload/finalize de scripts preserve a exigencia de `sparkit`.
