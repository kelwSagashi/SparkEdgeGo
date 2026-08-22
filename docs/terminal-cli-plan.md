# SparkEdge Terminal CLI Plan

Este documento define o plano de implementação de uma camada de operacao por terminal para o SparkEdge Go.

Motivacao principal:

- em Raspberry Pi 3 e outros devices remotos, o acesso via RaspConnect/RDP fica lento demais;
- o sistema precisa continuar administravel mesmo sem WebUI responsiva;
- a operacao local deve ser possivel diretamente pelo terminal, com ou sem rede;
- o CLI precisa cobrir as mesmas tarefas administrativas mais comuns que hoje sao feitas pela WebUI.

Este guia serve como referencia para a implementacao incremental da nova interface de manutencao do Edge.

## 1. Objetivo

Criar uma experiencia de linha de comando para o SparkEdge capaz de:

- criar e autenticar usuarios;
- fazer login e logout;
- criar, listar, visualizar, editar e remover scripts;
- instalar scripts via ZIP ou criar scripts por arquivos;
- criar e gerenciar credenciais;
- criar e gerenciar servidores;
- descobrir recursos e operacoes;
- criar e editar instancias;
- iniciar, parar e inspecionar execucoes;
- visualizar logs, status e diagnosticos;
- aplicar configuracoes locais;
- executar rotinas de update assistido e rollback;
- operar em modo completamente local, sem dependencia do Cloud para funcionar.

## 2. Escopo funcional

### 2.1 Deve ficar disponivel no terminal

- autenticacao local;
- gerenciamento de usuarios;
- gerenciamento de scripts;
- gerenciamento de credenciais;
- gerenciamento de servidores e providers;
- gerenciamento de devices;
- gerenciamento de instancias;
- consulta de execucoes e historicos;
- leitura e edicao de configuracao;
- update assistido;
- verificacoes de saude;
- exportacao e importacao de bundles.

### 2.2 Nao deve ser obrigatorio na primeira fase

- interface TUI completa estilo dashboard;
- edicao visual avancada com drag and drop;
- configuracoes complexas em modo assistido dentro do terminal;
- todos os fluxos da WebUI reescritos no primeiro corte.

## 3. Principios de arquitetura

### 3.1 Edge independente

O SparkEdge continua sendo um sistema independente.

Isso significa:

- ele deve continuar operando sem SparkCloud;
- o CLI precisa funcionar mesmo em ambientes offline;
- Cloud e apenas uma integracao opcional;
- configuracoes locais continuam sendo a fonte primaria para o Edge.

### 3.2 Paridade com a WebUI

O CLI deve refletir as mesmas capacidades da WebUI, mas com um fluxo mais rapido para ambientes lentos.

Na pratica:

- tudo que e essencial para administrar o Edge precisa existir no CLI;
- a WebUI continua existindo para cenarios visuais e operacao confortavel;
- o CLI precisa ser bom o suficiente para ser o caminho preferido em Raspberry e acesso remoto lento.

### 3.3 Perfil low-resource

O alvo principal inclui Raspberry Pi 3, armv7 e outros ambientes com pouca RAM e CPU.

Por isso:

- evitar dependencias pesadas;
- preferir comandos curtos e preditiveis;
- evitar TUI complexa no primeiro momento;
- manter execucao rapida e consumo de memoria baixo;
- privilegiar saidas simples em texto, JSON e formatos estruturados.

## 4. Decisoes tecnicas propostas

### 4.1 Estrutura do CLI

Usar uma aplicacao principal em Go com subcomandos separados.

Proposta de arvore:

```text
sparkedge
  auth
    login
    logout
    whoami
  users
    list
    create
    update
    delete
  scripts
    list
    show
    create
    edit
    install
    uninstall
    run
    playground
  credentials
    list
    create
    update
    delete
    test
  servers
    list
    create
    update
    delete
    discover
    resources
  devices
    list
    create
    update
    delete
  instances
    list
    create
    update
    delete
    run
    stop
    logs
    history
  config
    show
    edit
    validate
  update
    check
    apply
    rollback
  shell
  health
  version
```

### 4.2 Bibliotecas e padroes

Proposta inicial:

- `cobra` para roteamento de comandos e flags;
- `viper` ou `koanf` para leitura de configuracao;
- `fatih/color` ou equivalente simples para saida colorida;
- `json` padrao do Go para saida estruturada;
- bibliotecas leves de prompt somente se necessario;
- evitar TUI pesada antes de validar o valor real no campo.

### 4.3 Modo de operacao

O CLI precisa ter dois modos:

1. **modo nao interativo**
   - ideal para scripts, automacao, CI, SSH e uso remoto;
   - recebe flags e argumentos completos;
   - retorna JSON ou tabela simples;
   - falha com codigo de saida adequado.

2. **modo interativo**
   - usado quando o usuario quer fluxos guiados;
   - pode sugerir proximos passos;
   - pode pedir confirmacao antes de alterar dados sensiveis.

### 4.4 Shell interativo tipo bash

A experiencia ideal para essa fase tambem pode incluir um shell persistente, parecido com `bash`, `python` ou `docker exec`.

Nesse modo:

- o usuario entra em uma sessao do SparkEdge com `sparkedge shell`;
- o prompt vira uma interface de operacao dentro do proprio sistema;
- os comandos ficam disponiveis sem precisar sair e voltar para cada acao;
- o contexto atual pode ser exibido no prompt, como edge, usuario, projeto ou instancia;
- o shell pode oferecer auto-complete, ajuda contextual e historico de comandos.

Esse formato e importante porque transforma o CLI em um ambiente operacional, nao apenas em uma colecao de comandos isolados.

### 4.5 Persistencia e integracao

O CLI nao deve duplicar regras de negocio.

Em vez disso:

- comandos chamam os mesmos servicos internos do backend Go;
- validacoes ficam centralizadas nos modulos de dominio;
- o CLI e apenas uma camada de interface;
- qualquer acao feita no CLI deve refletir no mesmo SQLite, mesmos providers e mesmos scripts que a WebUI usa.

### 4.6 Formato de saida

O CLI precisa suportar:

- saida humana resumida;
- saida JSON para automacao;
- saida silenciosa quando necessario;
- codigo de erro consistente.

Recomendacao:

- default: tabela ou texto curto;
- `--json`: retorno estruturado;
- `--quiet`: reduzir ruido;
- `--verbose`: diagnosticos extras.

## 5. Fluxos prioritarios

### 5.1 Autenticacao e usuarios

Objetivo:

- criar usuario local;
- fazer login no Edge;
- guardar sessao local com JWT;
- permitir troca de usuario sem depender da WebUI.

Comandos sugeridos:

- `sparkedge auth login`
- `sparkedge auth logout`
- `sparkedge auth whoami`
- `sparkedge users create`
- `sparkedge users list`

### 5.2 Scripts

Objetivo:

- registrar script por ZIP ou diretorio;
- criar script textual sem compactar;
- editar script existente;
- ver detalhes, README e bundle;
- executar playground local no terminal;
- testar entrada e saida do script.

Comandos sugeridos:

- `sparkedge scripts install <arquivo.zip>`
- `sparkedge scripts create`
- `sparkedge scripts edit <id>`
- `sparkedge scripts show <id>`
- `sparkedge scripts run <id>`
- `sparkedge scripts playground <id>`

Pontos tecnicos:

- reutilizar o fluxo de venv por script;
- instalar requirements automaticamente;
- usar Sparkit no processamento da execucao;
- padronizar leitura de stdout e stderr;
- manter historico de versoes do bundle.

### 5.3 Credenciais e providers

Objetivo:

- criar e testar credenciais;
- descobrir providers e adaptadores disponiveis;
- permitir configuracao por terminal sem abrir a WebUI.

Comandos sugeridos:

- `sparkedge credentials create`
- `sparkedge credentials test`
- `sparkedge providers list`
- `sparkedge providers inspect <tipo>`

### 5.4 Servidores

Objetivo:

- criar servidor;
- selecionar credential;
- testar conexao;
- descobrir recursos e operacoes;
- visualizar schemas retornados pelo provider.

Comandos sugeridos:

- `sparkedge servers create`
- `sparkedge servers update`
- `sparkedge servers discover <id>`
- `sparkedge servers resources <id>`

### 5.5 Instancias

Objetivo:

- criar instancia completa;
- escolher script, device, trigger e destinos;
- editar payload e mapeamento;
- iniciar ou inspecionar execucoes;
- depurar comportamento sem depender da WebUI.

Comandos sugeridos:

- `sparkedge instances create`
- `sparkedge instances edit <id>`
- `sparkedge instances run <id>`
- `sparkedge instances logs <id>`
- `sparkedge instances history <id>`

### 5.6 Configuracao local

Objetivo:

- ler e editar `config.yml`;
- validar parametros do sistema;
- registrar origem de cada valor: default, env ou yaml.

Comandos sugeridos:

- `sparkedge config show`
- `sparkedge config edit`
- `sparkedge config validate`

### 5.7 Update assistido

Objetivo:

- verificar novas versoes;
- baixar pacote;
- aplicar update com backup;
- reiniciar ou orientar reinicio;
- permitir rollback caso algo falhe.

Comandos sugeridos:

- `sparkedge update check`
- `sparkedge update apply`
- `sparkedge update rollback`

## 6. Estrutura interna sugerida

Uma separacao pratica para o codigo Go:

```text
cmd/
  sparkedge/
    main.go
internal/
  app/
    bootstrap
    lifecycle
  auth/
  cli/
    commands
    prompts
    output
  config/
  credentials/
  devices/
  instances/
  scripts/
  servers/
  providers/
  updates/
  users/
  web/
  domain/
  storage/
```

Ideia:

- `cmd` apenas inicia a aplicacao;
- `internal/cli` concentra comandos e parsing;
- `internal/domain` e `internal/storage` concentram regras e persistencia;
- `internal/web` continua servindo a WebUI;
- `internal/app` une os modulos.

## 7. Fluxo de implementacao sugerido

### Fase 1 - base do CLI

- criar entrypoint de comando;
- criar comandos base `version`, `health`, `config show`;
- padronizar saida JSON e texto;
- integrar com config local.

### Fase 2 - autenticacao e usuarios

- login/logout;
- leitura de JWT local;
- status da sessao;
- CRUD minimo de usuarios.

### Fase 3 - scripts

- listagem;
- install via ZIP;
- create/edit por arquivos;
- readme e playground no terminal;
- historico de versoes.

### Fase 4 - credenciais, providers e servidores

- CRUD de credenciais;
- teste de conexao;
- descoberta de recursos;
- inspeccao dos schemas.

### Fase 5 - instancias

- create/edit via CLI;
- escolha de script e device;
- mapeamento de payload e destino;
- execucao e logs.

### Fase 6 - operacao e manutencao

- update assistido;
- rollback;
- backup;
- diagnostico;
- logs e health.

## 8. Regras de experiencia do usuario

- comandos precisam ser curtos;
- erros devem explicar o que falhou e como corrigir;
- entradas repetidas devem poder ser automatizadas por flags;
- quando houver risco, pedir confirmacao antes de alterar;
- se a operacao depender de rede, deixar isso claro;
- se houver fallback local, explicitar o que sera mantido em cache.

## 9. Pontos de integracao com SparkCloud

Mesmo com o foco em independencia, o CLI precisa respeitar o ecossistema SparkCloud quando houver conexao.

Pontos esperados:

- autenticacao com JWT;
- sincronizacao de metadados quando permitido;
- registro de edges;
- consulta de estado da instacia;
- envio de eventos e telemetria;
- operacao degradada quando o Cloud estiver offline.

## 10. Riscos e cuidados

- nao misturar logica de CLI com logica de dominio;
- nao duplicar validacao em varios lugares;
- nao obrigar modo interativo em ambientes sem TTY;
- nao criar dependencia pesada que prejudique armv7;
- nao fazer o CLI depender da WebUI;
- nao acoplar update assistido a internet o tempo todo.

## 11. Critérios de sucesso

O bloco de CLI pode ser considerado util quando:

- um Raspberry Pi 3 consegue operar sem abrir a WebUI;
- scripts podem ser instalados e testados por terminal;
- credenciais, servidores e instancias podem ser criados por CLI;
- o sistema continua independente do SparkCloud;
- comandos podem ser usados via SSH com baixa latencia;
- o mesmo backend Go atende CLI e WebUI;
- logs e diagnosticos ficam acessiveis em texto e JSON.

## 12. Pendencias para discutir

Ainda precisamos definir:

- o nome final do binario e dos subcomandos;
- se o modo interativo vai usar prompts simples ou uma TUI;
- qual formato padrao de saida sera preferido;
- como o CLI vai autenticar sem expor segredos;
- quais comandos devem exigir confirmacao;
- qual o minimo de comandos para a primeira entrega;
- se `sparkedge` e o binario unico ou se existira `sparkedgectl`.

## 13. Proximo passo

O proximo passo recomendado e transformar este plano em backlog executavel:

1. definir a arvore final dos comandos;
2. mapear cada comando para os servicos internos existentes;
3. decidir o formato de saida;
4. iniciar pela autenticacao e pelos comandos de leitura;
5. expandir para scripts e instancias.
