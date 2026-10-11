# SparkEdge

O SparkEdge e o agente de computacao de borda do ecossistema Spark. Ele e instalado proximo aos sistemas de energia - em um Raspberry Pi ou outro computador de baixo consumo - para conectar equipamentos locais a plataforma Spark sem depender de uma conexao continua com a internet.

Em vez de apenas encaminhar dados, o SparkEdge orquestra o ciclo de telemetria no local onde ela e produzida: executa coletas periodicas, integra scripts e drivers para inversores, controladores, baterias e sensores, valida e persiste os registros em SQLite, e envia os dados para a nuvem quando ha conectividade. Dessa forma, a coleta continua durante falhas de rede e a sincronizacao posterior preserva o historico para analise na plataforma web.

Essa abordagem de borda reduz a dependencia de equipamentos proprietarios e permite que cada unidade use os protocolos, colunas e metricas que realmente estao disponiveis. O SparkEdge tambem recebe configuracoes e comandos pelo canal MQTT, mantendo a comunicacao com o Spark desacoplada da leitura dos equipamentos em campo.

## Como os dados circulam

1. **Coleta na borda:** scripts Python com Sparkit, drivers e sensores leem os equipamentos no local.
2. **Orquestracao local:** o agente agenda execucoes, organiza instancias e registra os resultados no banco SQLite local.
3. **Resiliencia de conectividade:** os dados permanecem na unidade enquanto a rede estiver indisponivel; nao ha perda por depender de uma requisicao imediata.
4. **Sincronizacao e comunicacao:** quando a conexao esta disponivel, o SparkEdge envia a telemetria de forma confiavel ao Spark e mantem MQTT para provisionamento, configuracoes e eventos.
5. **Visualizacao e operacao:** a plataforma Spark transforma os registros sincronizados em tabelas, metricas, alertas e paineis para acompanhamento remoto.

## Principais capacidades

- **Coleta aberta e extensivel:** integra scripts Python, drivers e sensores para equipamentos de diferentes fabricantes, sem limitar a unidade a uma linha proprietaria.
- **Automacao local:** executa rotinas periodicas e organiza instancias de coleta, mesmo quando nao ha acesso externo ao dispositivo.
- **Persistencia offline-first:** usa SQLite como fonte local de verdade e sincroniza os registros posteriormente, quando houver conexao.
- **Comunicacao com a nuvem:** usa MQTT e servicos de sincronizacao para receber configuracoes e encaminhar telemetria ao Spark.
- **Operacao em campo:** oferece WebUI, API REST e CLI local para configuracao, diagnostico e administracao do agente.

# SparkEdge Go

O SparkEdge Go e a base atual do agente, substituindo o sistema anterior em TypeScript, [SparkEdge](https://github.com/kelwSagashi/SparkEdge). Go permite distribuir binarios multiplataforma com uma operacao mais simples em dispositivos de borda, especialmente Raspberry Pi e computadores compactos instalados em campo.

Este repositorio preserva o frontend Vite e implementa os componentes necessarios para a orquestracao local de dados:

- banco principal local em SQLite;
- manipulacao do SQLite por ORM, evitando SQL manual nos repositories;
- execucao de scripts Python padronizada por Sparkit;
- MQTT via EMQX;
- providers e drivers plugaveis para destinos externos;
- servidor REST modular;
- CLI local;
- motor de execucao de instancias como componente de primeira classe.

As versoes do SparkEdge Go a partir da 0.0.0 sucedem a versao 0.1.2 do [SparkEdge](https://github.com/kelwSagashi/SparkEdge), que foi descontinuado.


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
O runbook operacional de release, instalacao em campo, Raspberry e update assistido esta em [docs/operations-runbook.md](./docs/operations-runbook.md).
O plano da nova operacao via terminal para administrar o Edge sem depender da WebUI esta em [docs/terminal-cli-plan.md](./docs/terminal-cli-plan.md).

# Inicialização

Baixe um binario compativel com a sua maquina, descompacte, inicie o executavel e abra a webui do SparkEdge em http://localhost:3009

## Raspberry Pi 3

Para Raspberry Pi 3 com sistema 32 bits, use o pacote `linux-armv7`.

Passo rapido:

1. baixar a release `linux-armv7`;
2. extrair o `.zip` no Raspberry;
3. entrar na pasta `sparkEdge`;
4. dar permissao de execucao ao binario;
5. iniciar o SparkEdge de dentro dessa pasta.

Comandos:

```bash
cd ~/sparkEdge
chmod +x sparkedge
./sparkedge
```

Depois, abrir:

- `http://localhost:3009/` no proprio Raspberry;
- ou `http://IP_DO_RASPBERRY:3009/` a partir de outra maquina da rede.

Lembrete importante:

- rode o binario dentro da pasta `sparkEdge`;
- mantenha `config.yml` ao lado do executavel;
- mantenha `webui/dist` ao lado do executavel;
- o `sparkedge.db` sera criado ou reutilizado nesse mesmo local.
