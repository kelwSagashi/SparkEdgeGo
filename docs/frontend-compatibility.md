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

Quando trouxermos o frontend para o repositorio Go, este e um dos ajustes que precisam ser feitos.

### 2. Base URLs e modo de execucao

Hoje o frontend usa:

- `fetch("/api/...")` em alguns pontos
- `axios` com `baseURL: "/api"` em outros

Isso e bom para servir o frontend pelo mesmo host do backend Go, mas ao sair do monorepo TypeScript precisamos revisar:

- pipeline de build do Vite
- localizacao final dos assets estaticos
- variaveis de ambiente do frontend
- paths relativos usados no desenvolvimento e na distribuicao

### 3. Tipos importados do monorepo TypeScript

O frontend ainda importa varios tipos de `spark-edge-db/src/types`.

Quando ele for separado do monorepo TypeScript, precisaremos decidir entre:

- gerar tipos locais para o frontend
- copiar contratos necessarios para um pacote compartilhado novo
- substituir imports por tipos proprios do frontend

## Recomendacao para a fase de integracao

Antes de mover o frontend:

1. consolidar o cliente HTTP do frontend em uma unica camada
2. alinhar as tipagens das respostas reais do backend Go
3. decidir como os tipos compartilhados vao sair do monorepo TypeScript
4. so depois conectar o build do Vite ao servidor Go
