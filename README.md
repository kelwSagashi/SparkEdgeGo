# SparkEdge

O SparkEdge é uma plataforma baseada de automação para monitoramento com node, visando sistemas de energia fotovoltáicos on-grid e off-grid. A plataforma oferece e permite integração, automação e execução de código para coleta de dados em sistemas de energia permitindo maior controle na automação e envio de dados.

# Principais capacidades
- **Integração com Código Python**: É possivel usar código em python para coleta de dados e adicionar como uma instancia.
- **Controle Total**: Você define o que fazer com os dados de monitoramento que foram coletados.
- **Serviço local**: Permite rodar em sistemas mais criticos tendo flexibilidade em conseguir armazenar dados e enviar quando houver conexão.


# SparkEdge Go

Nova base do SparkEdge em Go subistituido o sistema antigo construido em typescript [SparkEdge](https://github.com/kelwSagashi/SparkEdge).

Este repositorio nasce para substituir o backend TypeScript do SparkEdge, preservando o frontend Vite e mantendo os conceitos atuais do sistema:

- banco principal local em SQLite;
- manipulacao do SQLite por ORM, evitando SQL manual nos repositories;
- execucao de scripts Python padronizada por Sparkit;
- MQTT via EMQX;
- providers e drivers plugaveis para destinos externos;
- servidor REST modular;
- CLI local;
- motor de execucao de instancias como componente de primeira classe.

Essa versão vem com o objetivo de atender requisitos de multiplataforma que são de extrema importancia para esse projeto.


A versão 0.0.0 do Spark Edge Go é equivalente a versão 0.1.2 do [SparkEdge](https://github.com/kelwSagashi/SparkEdge) (descontinuado). Então todas as versões do spark Edge Go a partir da 0.0.0 passam a estar a frente da ultima versão lançada da versão descontinuada.


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

# Inicialização

Baixe um binario compativel com a sua maquina, descompacte, inicie o executavel e abra a webui do SparkEdge em http://localhost:3009


# O que significa SparkEdge?
O nome SparkEdge simboliza o "faísca" (spark) da energia e o foco no processamento de dados na "borda" (edge) da rede, refletindo o foco da plataforma em processamento local e monitoramento eficiente de sistemas de energia, principalmente em regiões onde tenha baixa conexão.
