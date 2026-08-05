# SparkEdge Go

Nova base do SparkEdge em Go.

Este repositorio nasce para substituir o backend TypeScript do SparkEdge, preservando o frontend Vite e mantendo os conceitos atuais do sistema:

- banco principal local em SQLite;
- execucao de scripts Python padronizada por Sparkit;
- MQTT via EMQX;
- providers e drivers plugaveis para destinos externos;
- servidor REST modular;
- CLI local;
- motor de execucao de instancias como componente de primeira classe.

## Estado inicial

Esta primeira base e intencionalmente pequena. Ela cria a fundacao do repositorio Go antes de portar comportamento:

- `cmd/sparkedge-api`: entrada do servidor REST.
- `cmd/sparkedge-cli`: entrada da CLI.
- `internal/app`: composicao da aplicacao.
- `internal/domain`: entidades e contratos de dominio.
- `internal/httpapi`: servidor HTTP e padrao de resposta.
- `internal/sqlite`: fronteira do banco local.
- `internal/runtime`: motor de execucao de instancias.
- `internal/python/sparkit`: integracao com scripts Sparkit.
- `internal/providers`: registry de destinos externos.
- `internal/mqtt`: fronteira EMQX.

## Proximas ondas

1. Inventariar rotas e contratos do backend TypeScript.
2. Migrar schema SQLite e repositorios.
3. Portar o motor de execucao de instancias.
4. Portar providers/drivers.
5. Portar API REST e CLI.
6. Reapontar o frontend Vite.

