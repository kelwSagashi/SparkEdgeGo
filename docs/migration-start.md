# Inicio da migracao

Este documento registra o primeiro passo da migracao do SparkEdge para Go.

## Decisoes aplicadas

- A versao Go fica em um novo repositorio: `SparkEdgeGo`.
- O backend TypeScript existente permanece como referencia de comportamento.
- O banco principal continua sendo SQLite local.
- Sparkit e o contrato padronizado para scripts Python.
- EMQX e o broker MQTT esperado.
- Providers externos continuam plugaveis e nao substituem o SQLite.
- O motor de execucao de instancias e uma area propria do backend.

## Primeiro esqueleto

A primeira estrutura foi criada com biblioteca padrao de Go para evitar dependencias externas antes do inventario completo:

- servidor HTTP com health check;
- CLI com comandos `start` e `status`;
- fronteiras para SQLite, EMQX, Sparkit, providers e runtime;
- entidades iniciais para instancia, script e execucao.

## Proximo trabalho recomendado

1. Portar o schema SQLite atual para migracoes Go.
2. Criar o inventario de controllers e endpoints.
3. Portar o contrato do `InstanceRunnerService`.
4. Implementar executor Sparkit real com `--schema`, `--input` e `--input-file`.
5. Criar os primeiros providers: HTTP e Supabase.

