# Compatibilidade do Frontend Vite

Este documento registra os pontos encontrados ao revisar o frontend atual antes da futura integracao com o servidor Go.

## Rotas legadas encontradas no frontend

Durante a leitura de `packages/frontend/src`, o frontend ainda referencia alguns endpoints antigos:

- `GET /api/scripts/downloaded`
- `GET /api/cli/mqtt-config`

Para reduzir atrito na futura migracao do frontend, a versao Go agora expoe ambos:

- `GET /api/scripts/downloaded` como alias de `GET /api/scripts`
- `GET /api/cli/mqtt-config` com payload simples:

```json
{
  "broker_url": "mqtt://...",
  "username": "spark-user-edge-...",
  "has_password": true,
  "password": "..."
}
```

## O que ja foi resolvido na migracao Go

- o frontend foi movido para `frontend/` dentro do repositorio Go;
- os tipos antes puxados do monorepo TypeScript passaram a existir localmente em `src/types`;
- o build do Vite agora gera `frontend/dist`;
- o binario Go ja consegue servir `GET /`, `GET /api/health` e os assets do Vite no mesmo host;
- o cliente HTTP do frontend foi parcialmente consolidado em uma instancia central de API;
- o contrato de `GET /api/instances/{id}` foi alinhado para o formato `{ instance, destinations }`.

## Pontos de atencao ainda abertos

### 1. Tipagem duplicada de instancias no frontend

Existem duas camadas consumindo `/api/instances/{id}`:

- `src/server/server.service.ts`
- `src/rest-api-client/instances.service.ts`

A resposta real do backend Go para `GET /api/instances/{id}` e:

```json
{
  "data": {
    "instance": { "...": "..." },
    "destinations": [ ... ]
  },
  "error": null
}
```

`server.service.ts` ja espera esse formato. `instances.service.ts` ainda tipa essa rota como se retornasse apenas `InstanceReturningValues`.

Este ajuste ja foi aplicado na camada `rest-api-client`, mas ainda vale revisar chamadas antigas que passam por `server.service.ts`.

### 2. Camada HTTP ainda duplicada

Ainda existem duas abordagens convivendo no frontend:

- `src/server/server.service.ts`
- `src/rest-api-client/*`

Hoje ambas funcionam sobre o mesmo host `/api`, mas ainda e desejavel consolidar isso em uma unica camada para reduzir divergencias de tipagem e manutencao.

### 3. Tipos locais ainda estao amplos

Para desacoplar a migracao do monorepo TypeScript, varios contratos foram internalizados no frontend com tipagem propositalmente mais permissiva.

O proximo refinamento natural e endurecer esses tipos conforme os contratos do backend Go forem estabilizando.

## Recomendacao para a fase de integracao

1. validar as principais telas do frontend ja servido pelo binario Go
2. consolidar a camada HTTP em uma unica estrategia
3. endurecer as tipagens locais mais permissivas
4. seguir para empacotamento final e validacoes com integracoes reais
