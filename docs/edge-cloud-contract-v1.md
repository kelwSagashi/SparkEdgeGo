# Edge <-> Cloud Contract v1

Data: 2026-08-13

## Objetivo

Definir o contrato minimo entre `SparkEdgeGo` e `SparkAPI` para presenca, telemetria operacional, metadata do edge e comandos remotos.

Este contrato deve:

- preservar compatibilidade com payloads legados quando possivel;
- permitir evolucao progressiva;
- ser leve o suficiente para ambientes de baixa conectividade;
- nao acoplar o runtime local do edge ao cloud.

## Principios

- o `SparkEdge` continua operando sem cloud;
- o `SparkCloud` consome sinais do edge quando conectividade existir;
- eventos relevantes devem ser idempotentes;
- MQTT e o canal principal para presenca e comando;
- HTTP permanece para onboarding, pairing, auth e sincronizacoes futuras em lote.

## Topicos MQTT

Base:

- `spark/{edge_id}/status`
- `spark/{edge_id}/heartbeat`
- `spark/{edge_id}/stats`
- `spark/{edge_id}/meta`
- `spark/{edge_id}/context`
- `spark/{edge_id}/commands`
- `spark/{edge_id}/response`
- `spark/{edge_id}/metrics`
- `spark/{edge_id}/logs`

## Envelope padrao

Payloads estruturados novos devem carregar:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "heartbeat",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z"
}
```

### Campos obrigatorios

- `schema_version`
- `message_id`
- `type`
- `edge_id`
- `occurred_at`

### Campos de compatibilidade

- `timestamp` continua aceito no cloud
- `ts` pode ser mantido em heartbeat quando util para compatibilidade

## Status

Topico:

- `spark/{edge_id}/status`

Payload recomendado:

- string simples: `online` ou `offline`

Observacao:

- manter string simples favorece Last Will e parsing leve
- estados mais ricos vao em `heartbeat`

## Heartbeat v2

Topico:

- `spark/{edge_id}/heartbeat`

Payload:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "heartbeat",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z",
  "ts": 1786622400,
  "status": "online",
  "connectivity": {
    "mqtt_connected": true,
    "cloud_connected": null
  },
  "runtime": {
    "uptime_seconds": 1234,
    "goroutines": 17
  }
}
```

Objetivo:

- manter o edge vivo na visao da malha
- carregar estado operacional leve

## Stats v2

Topico:

- `spark/{edge_id}/stats`

Payload:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "stats",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z",
  "data": {
    "cpu_pct": 0,
    "memory_mb": 42.1,
    "disk_pct": 0,
    "uptime_seconds": 1234,
    "goroutines": 17,
    "network": {
      "latency_ms": null
    }
  }
}
```

Objetivo:

- carregar snapshot operacional um pouco mais rico que heartbeat
- ficar apto para cache no cloud

## Meta v2

Topico:

- `spark/{edge_id}/meta`

Payload:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "meta",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z",
  "edge_name": "Edge Planta A",
  "description": "Borda da usina principal",
  "lat": "-12.1234",
  "lng": "-38.9987",
  "location_source": "manual",
  "tags": ["solar", "bahia"],
  "os": "linux",
  "os_version": "linux/arm64",
  "edge_version": "v0.2.0",
  "hardware": "arm64",
  "environment": "production"
}
```

Objetivo:

- sincronizar inventario do edge
- suportar localizacao e metadata da malha

## Context v1

Topico:

- `spark/{edge_id}/context`

Payload:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "context",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z",
  "local_user": {
    "id": "user-1",
    "name": "Local Operator",
    "email": "local@example.com"
  }
}
```

## Command request

Topico:

- `spark/{edge_id}/commands`

Payload minimo:

```json
{
  "command_id": "cmd-123",
  "type": "get_stats",
  "payload": {},
  "context": {
    "requested_by": "cloud-user"
  }
}
```

## Command response

Topico:

- `spark/{edge_id}/response`

Payload minimo:

```json
{
  "schema_version": "edge-cloud.v1",
  "message_id": "uuid",
  "type": "command_response",
  "edge_id": "edge-123",
  "occurred_at": "2026-08-13T12:00:00Z",
  "timestamp": "2026-08-13T12:00:00Z",
  "command_id": "cmd-123",
  "status": "done",
  "result": {},
  "error": null
}
```

### Estados de resposta de comando

Sequencia recomendada:

1. `received`
2. `running`
3. `done` ou `error`

Objetivo:

- separar publicacao no cloud de recebimento real pelo edge;
- diferenciar fila/entrega de execucao;
- permitir auditoria de timeout e comandos travados.

## Compatibilidade

O cloud deve aceitar:

- status como string simples
- heartbeat antigo com apenas `edge_id` e `ts`
- response antiga com `command_id`, `status`, `result`
- meta sem `schema_version`

O edge deve publicar no formato novo sempre que possivel.

## Idempotencia

Nesta primeira fase:

- `message_id` deve ser unico por publicacao estruturada
- `command_id` continua sendo o identificador funcional do ciclo de comando

Na fase seguinte:

- o cloud deve deduplicar por `edge_id + message_id` quando persistir eventos sincronizaveis

## Frequencias sugeridas

- `status`: na conexao, desconexao e via LWT
- `heartbeat`: 30s
- `stats`: 120s
- `meta`: no reconnect, no onboarding sync e quando metadata local mudar
- `context`: quando usuario local relevante mudar ou ao conectar

## Fora do escopo desta versao

- fila geral de sincronizacao HTTP em lote
- batch de telemetria
- compressao
- deduplicacao persistida por `message_id`
- politicas de prioridade de sync
