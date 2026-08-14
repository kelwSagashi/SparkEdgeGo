# SparkEdge + SparkCloud Roadmap

Data de elaboracao: 2026-08-13

## Objetivo

Definir a proxima fase de evolucao do ecossistema Spark, alinhando:

- `SparkEdgeGo`: runtime local autonomo, resiliente e multiplataforma.
- `SparkAPI`: coordenacao, inventario, telemetria, pareamento e comando remoto.
- `SparkEdge` TypeScript atual: referencia funcional de comportamentos legados e contratos ja usados pelo frontend local.

O objetivo nao e transformar o SparkEdge em um cliente dependente do SparkCloud. O objetivo e consolidar:

- `SparkEdge` como no operacional autonomo da borda.
- `SparkCloud` como camada opcional de monitoramento, coordenacao e gestao da malha.

## Leitura atual do produto

Com base na base atual:

- O `SparkEdgeGo` ja executa scripts localmente, resolve mappings, entrega em destinos externos, faz fallback local e usa MQTT/EMQX para integracao e presenca.
- O `SparkAPI` ja mantem inventario de edges, credenciais MQTT, metrics, activities, commands, pairing tokens e eventos de conectividade.
- O frontend legado do `SparkEdge` comprova que o produto foi desenhado como uma plataforma de automacao e integracao na borda, com baixa dependencia de conectividade e forte configurabilidade de destinos.

Conclusao:

- o `SparkEdge` e a fonte primaria de operacao;
- o `SparkCloud` e a fonte secundaria de visibilidade global da malha;
- a sincronizacao precisa ser assincrona, idempotente e tolerante a falhas.

## Repositorios impactados

### SparkEdgeGo

Path: `C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdgeGo`

Areas ja relevantes:

- `internal/runtime`
- `internal/mqtt`
- `internal/edge`
- `internal/httpapi`
- `internal/sqlite`
- `internal/updater`
- `webui`

### SparkAPI

Path: `C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkAPI`

Areas ja relevantes:

- `src/modules/mqtt`
- `src/modules/edges`
- `src/modules/commands`
- `src/modules/pairing`
- `src/modules/activities`
- `src/modules/mqtt-broker`
- `prisma/schema.prisma`

### SparkEdge TypeScript legado

Path: `C:\Users\kelwp\OneDrive\Documentos\bolsa piape\monitor-manager\SparkCloud\SparkEdge`

Uso nesta fase:

- referencia de fluxos legados;
- consulta de comportamentos ainda nao migrados;
- comparacao de UX e contratos locais de runtime.

## Arquitetura-alvo

### Papel do SparkEdge

Responsabilidades:

- executar scripts e instancias localmente;
- processar dados e automacoes sem internet;
- persistir configuracoes, execucoes, filas e historico localmente;
- entregar dados para destinos configurados;
- manter fila de sincronizacao opcional com o SparkCloud;
- aceitar comandos remotos quando conectado;
- operar em modo degradado quando a conectividade cair.

### Papel do SparkCloud

Responsabilidades:

- inventariar e identificar edges da malha;
- manter estado resumido e historico de conectividade;
- receber telemetria operacional do edge;
- receber dados replicados quando o edge desejar sincronizar;
- publicar comandos remotos e acompanhar ACK/resultado;
- exibir mapa, presenca, saude e distribuicao da malha;
- agregar observabilidade global sem bloquear a operacao local.

### Principio tecnico

Prioridade de verdade:

1. operacao local;
2. persistencia local;
3. sincronizacao com cloud quando disponivel.

## Macro metas

### Meta 1 - Contrato Edge <-> Cloud

Objetivo:

Definir um contrato estavel entre `SparkEdgeGo` e `SparkAPI` para:

- onboarding/pairing;
- presenca;
- telemetria operacional do edge;
- sincronizacao de dados;
- sincronizacao de configuracao;
- comandos remotos;
- ack e resultado de execucao.

Como fazer:

- levantar os payloads atuais de `pairing`, `commands`, `mqtt topics`, `metrics` e `activities`;
- versionar os contratos em documento unico;
- padronizar `snake_case` no fio e adaptar internamente se necessario;
- introduzir `schema_version` e `message_id` nos eventos relevantes;
- distinguir claramente `telemetry data`, `edge status`, `edge metadata`, `command response` e `sync event`.

Decisoes tecnicas:

- MQTT continua sendo o canal principal para presenca e comando em tempo quase real;
- HTTP continua sendo o canal principal para onboarding, consulta e eventualmente sincronizacao em lote;
- payloads devem ser idempotentes e carregarem `edge_id`, `message_id`, `occurred_at` e `schema_version`;
- o cloud nao deve depender de conhecer o schema completo de cada script para aceitar eventos.

Entregaveis:

- documento de contrato v1;
- tabela de topicos MQTT;
- tabela de endpoints HTTP;
- tabela de eventos e seus campos obrigatorios.

### Meta 2 - Presenca e telemetria operacional da malha

Objetivo:

Permitir que o SparkCloud monitore a saude real dos edges da malha.

Como fazer:

- enriquecer heartbeat do `SparkEdgeGo`;
- publicar status com mais semantica que apenas `online/offline`;
- publicar estatisticas locais em topico dedicado;
- atualizar `SparkAPI` para persistir os sinais mais importantes sem explodir volume.

Campos sugeridos:

- `edge_id`
- `status`: `online`, `degraded`, `offline`, `maintenance`
- `last_seen_at`
- `edge_version`
- `os`, `os_version`, `hardware`
- `uptime_seconds`
- `cpu_pct`
- `memory_pct`
- `disk_pct`
- `queue_sizes`
- `oldest_pending_age_seconds`
- `active_instances`
- `local_user`
- `lat`, `lng`, `location_source`

Decisoes tecnicas:

- heartbeat leve e frequente;
- stats mais pesadas em intervalo maior;
- manter MQTT para presenca;
- considerar snapshot resumido no banco do cloud e historico apenas para eventos relevantes.

Entregaveis:

- payload de heartbeat v2;
- payload de stats v1;
- atualizacao do parser MQTT em `SparkAPI`;
- adaptacao do modelo `Edge` para status operacional expandido.

### Meta 3 - Fila de sincronizacao com SparkCloud

Objetivo:

Criar sincronizacao assincrona e confiavel entre Edge e Cloud, sem acoplar a operacao local ao cloud.

Como fazer:

- criar fila local especifica para sync com cloud em `SparkEdgeGo`;
- separar essa fila do fallback de destinos comuns;
- armazenar itens com `type`, `priority`, `payload`, `message_id`, `status`, `retry_count`, `next_attempt_at`;
- enviar por lotes quando houver conectividade;
- usar ACK do cloud para confirmar;
- permitir replay seguro apos reconexao.

Itens tipicos da fila:

- status do edge;
- heartbeat consolidado;
- activities;
- snapshots de configuracao relevantes;
- telemetria marcada para replicacao;
- resultados de comando.

Decisoes tecnicas:

- fila persistente no SQLite local;
- retry com backoff exponencial;
- prioridade por tipo de evento;
- ACK por `message_id`;
- idempotencia obrigatoria no `SparkAPI`.

Entregaveis:

- nova tabela SQLite de sync;
- serviço de flush;
- endpoint ou topico de ingestao no cloud;
- status visual da fila na `webui`.

### Meta 4 - Controle remoto confiavel

Objetivo:

Fazer o comando remoto da malha ser auditavel, resiliente e seguro.

Como fazer:

- revisar o fluxo atual de `commands` no `SparkAPI`;
- padronizar estados: `pending`, `queued`, `delivered`, `running`, `done`, `error`, `timeout`, `expired`;
- fazer o edge publicar ACK de recebimento;
- fazer o edge publicar ACK de inicio e resultado final;
- permitir replay de comando enquanto ainda estiver valido;
- registrar trilha completa no cloud e localmente.

Decisoes tecnicas:

- comando remoto via MQTT em `spark/{edge_id}/commands`;
- resposta padronizada em `spark/{edge_id}/response`;
- `command_id` precisa existir ponta a ponta;
- prazo de validade no comando;
- payload de resultado separado de payload de logs.

Entregaveis:

- contrato de comando remoto v2;
- melhoria do `SparkAPI` em `commands` e parser MQTT;
- melhoria do `SparkEdgeGo` para ACK e resultado estruturado;
- tela de historico de comandos no cloud.

### Meta 5 - Descoberta, localizacao e inventario da malha

Objetivo:

Transformar edges em ativos observaveis de uma malha distribuida.

Como fazer:

- consolidar metadados de edge no onboarding e heartbeat;
- permitir atualizacao de localizacao pelo edge e pelo cloud;
- separar `location_source`: `manual`, `gps`, `network`, `cloud`;
- manter tags, ambiente, versao, hardware e vinculo organizacional.

Decisoes tecnicas:

- localizacao pode ser enviada no pairing inicial e atualizada depois;
- cloud pode sobrepor localizacao manualmente quando necessario;
- evitar gravar historico de localizacao em alta frequencia sem necessidade;
- manter `Edge` como inventario atual e `Activity`/historico separado para eventos.

Entregaveis:

- revisao do modelo de `Edge` no `SparkAPI`;
- payload de metadata do edge no `SparkEdgeGo`;
- ajuste do mapa e filtros de malha no frontend cloud quando essa camada for atacada.

### Meta 6 - Modo de baixa conectividade como recurso de produto

Objetivo:

Tornar a operacao em conectividade ruim um comportamento intencional e visivel.

Como fazer:

- introduzir um estado local de conectividade;
- reduzir heartbeat/stats quando necessario;
- priorizar trafego critico;
- fazer batch para sync com cloud;
- comprimir uploads HTTP quando aplicavel;
- limitar crescimento das filas;
- sinalizar degradacao na interface local e na nuvem.

Decisoes tecnicas:

- estados: `healthy`, `intermittent`, `offline`, `degraded`;
- filas com prioridade alta, media e baixa;
- politica de retencao por tipo de evento;
- batch apenas para eventos replicaveis, nao para comandos.

Entregaveis:

- modulo de policy de conectividade no `SparkEdgeGo`;
- configuracoes expostas em `config.yml` e UI avancada;
- indicadores de modo degradado na `webui` e no cloud.

### Meta 7 - Operacao soberana local

Objetivo:

Garantir que o SparkEdge continue forte mesmo sem SparkCloud.

Como fazer:

- reforcar backup e restore local;
- export/import de configuracao e scripts;
- ferramentas de diagnostico local;
- update por pacote local;
- painel local de saude do SQLite, filas e runtime.

Decisoes tecnicas:

- esse bloco e local-first e nao deve depender do cloud;
- a malha se beneficia porque edges remotos ficam recuperaveis sem backend.

Entregaveis:

- backup/restore local;
- pacote de exportacao/importacao;
- tela local de health operacional;
- retencao configuravel.

## Decisoes de contrato

### Canais

- MQTT:
  - presenca
  - heartbeat
  - stats
  - commands
  - responses
  - opcionalmente telemetry leve de tempo real

- HTTP:
  - auth e pairing
  - consulta de estado
  - sincronizacao em lote
  - download de configuracao futura
  - update metadata

### Idempotencia

Todo evento sincronizavel deve ter:

- `message_id`
- `edge_id`
- `schema_version`
- `occurred_at`
- `type`

No cloud, `message_id` precisa ser unico por `edge_id`.

### Prioridades

- alta:
  - status
  - command ack
  - command result
  - alertas criticos

- media:
  - heartbeat
  - stats
  - activities

- baixa:
  - telemetria replicada
  - snapshots auxiliares
  - dados analiticos historicos

### Retencao local

Regras minimas sugeridas:

- limite por fila;
- limite por idade;
- limite por tamanho total;
- descarte preferencial de itens de baixa prioridade;
- indicadores visuais antes de atingir estado critico.

## Mudancas por sistema

### SparkEdgeGo

Blocos principais:

1. `internal/cloudsync` novo modulo:
   - fila local
   - flush
   - politicas de prioridade
   - ack
   - replay

2. `internal/mqtt`:
   - heartbeat enriquecido
   - stats topic padronizado
   - ack de comando
   - resultado estruturado

3. `internal/edge`:
   - metadata do edge
   - presenca expandida
   - sync com cloud opcional

4. `internal/httpapi`:
   - health operacional
   - status da fila cloud
   - configuracao de conectividade

5. `internal/sqlite`:
   - tabela de cloud sync queue
   - tabela de sync log opcional

6. `webui`:
   - tela de conectividade
   - status do cloud sync
   - historico de comandos
   - saude local

### SparkAPI

Blocos principais:

1. `prisma/schema.prisma`:
   - ampliar modelo de sync, telemetria operacional e comando remoto
   - considerar tabela dedicada para eventos de edge se `metrics` ficar generica demais

2. `src/modules/mqtt`:
   - parser de heartbeat/stats/meta/response v2
   - persistencia seletiva
   - deduplicacao por `message_id`

3. `src/modules/commands`:
   - status expandido
   - validade e expiracao
   - historico auditavel

4. `src/modules/edges`:
   - endpoint de detalhes operacionais
   - endpoint de sync de estado, se optar por HTTP batch
   - listagem de malha com filtros de saude e conectividade

5. `src/modules/activities`:
   - tratar eventos de conectividade e degradacao

6. `src/modules/pairing`:
   - incluir `schema_version`
   - reforcar metadata inicial do edge

### SparkEdge legado

Uso recomendado:

- nao expandir funcionalidade nova aqui;
- usar apenas para consulta de paridade, comportamento antigo e comparacao de UX;
- migrar qualquer comportamento ainda util para `SparkEdgeGo` antes de investir em evolucao adicional no legado.

## Milestones propostas

### Milestone A - Contratos e telemetria basica da malha

Escopo:

- documento de contrato;
- heartbeat enriquecido;
- stats do edge;
- parser cloud atualizado;
- status expandido no inventario do cloud.

Resultado esperado:

O cloud enxerga melhor a malha e o edge passa a falar um protocolo mais claro.

### Milestone B - Comando remoto confiavel

Escopo:

- estados de comando revisados;
- ACK e resultado estruturado;
- auditoria local e cloud;
- expiracao e replay seguro.

Resultado esperado:

Comandos remotos deixam de ser apenas publish MQTT e passam a ser operacao rastreavel.

### Milestone C - Sync assincrono com cloud

Escopo:

- fila local cloud sync;
- retry e prioridade;
- ingestao idempotente no cloud;
- observabilidade da fila.

Resultado esperado:

O edge continua autonomo e o cloud recebe eventos quando houver conectividade.

### Milestone D - Baixa conectividade como recurso

Escopo:

- modo degradado;
- batch e politicas de envio;
- retencao local;
- painel local e cloud de saude da conectividade.

Resultado esperado:

A operacao remota fica economicamente e tecnicamente robusta.

### Milestone E - Soberania local reforcada

Escopo:

- backup/restore;
- export/import;
- health local;
- update por pacote.

Resultado esperado:

Cada edge se sustenta sozinho por longos periodos, mesmo sem cloud.

## Ordem recomendada de implementacao

1. Contratos e payloads
2. Heartbeat e stats v2
3. Fluxo de comandos com ACK
4. Fila de cloud sync
5. Ingestao idempotente no SparkAPI
6. UI local e cloud de conectividade
7. Politicas de baixa conexao
8. Backup/restore e operacao soberana local

## Riscos e cuidados

- nao misturar telemetria de edge com dados de negocio sem taxonomia clara;
- nao sobrecarregar o MQTT com payloads pesados;
- nao depender de HTTP online para algo critico de runtime;
- nao transformar a fila cloud em duplicata confusa do fallback de destinos;
- nao expandir o legado TypeScript alem do necessario;
- nao gravar historico em volume alto no cloud sem estrategia de retencao.

## Primeira fase sugerida

Para iniciar com menor risco e maior valor:

1. definir contrato Edge <-> Cloud v1;
2. implementar heartbeat/stats v2 no `SparkEdgeGo`;
3. adaptar `SparkAPI` para consumir esses payloads;
4. revisar fluxo de `commands` com ACK;
5. so depois abrir a fila de sincronizacao geral.

## Criterios de sucesso

- edge continua funcionando sem cloud;
- cloud enxerga a malha com qualidade;
- comando remoto fica confiavel e auditavel;
- reconexao nao duplica eventos;
- fila local nao cresce sem controle;
- degradacao de rede e visivel e administravel.
