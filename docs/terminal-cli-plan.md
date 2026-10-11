# SparkEdge Terminal CLI Plan

Este documento define o plano de implementação de uma camada de operacao por terminal para o SparkEdge Go.

Motivacao principal:

- em dispositivos remotos ou com poucos recursos, como Raspberry Pi 3, o acesso via RaspConnect/RDP fica lento demais;
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

O produto deve funcionar em qualquer plataforma suportada pelo binario Go. O Raspberry Pi 3, armv7 e outros ambientes com pouca RAM e CPU sao perfis de validacao de baixo recurso, nao o limite do produto.

Por isso:

- evitar dependencias pesadas;
- preferir comandos curtos e preditiveis;
- evitar TUI complexa no primeiro momento;
- manter execucao rapida e consumo de memoria baixo;
- privilegiar saidas simples em texto, JSON e formatos estruturados.

### 3.4 Perfis de plataforma

O planejamento deve considerar quatro perfis complementares:

- desktop: Windows, Linux e macOS em amd64 ou arm64, com uso local ou por terminal integrado;
- edge compacto: Linux arm64, armv7 ou armv6, com pouca memoria, armazenamento limitado e acesso remoto lento;
- servidor: Linux amd64 ou arm64, possivelmente operado por SSH, systemd ou container;
- automacao: ambientes CI/CD e scripts sem TTY, usando somente subcomandos e saida JSON.

Todos os perfis usam o mesmo dominio, banco SQLite, configuracao, autenticacao JWT e contrato de comandos. O que varia e o empacotamento, o processo de servico, o terminal disponivel e os limites de recursos.

### 3.5 Portabilidade esperada

O CLI deve:

- evitar caminhos, comandos de shell e sinais exclusivos de um sistema operacional;
- usar as APIs portaveis da biblioteca padrao Go para arquivos, processos, ambiente e terminal;
- detectar capacidades do terminal sem exigir cores, Unicode ou TTY;
- respeitar diretorios de configuracao e dados definidos pelo sistema operacional;
- aceitar `--data-dir`, `--config` e `--no-color` para operacao previsivel;
- manter o mesmo comportamento quando executado diretamente, como servico ou por SSH;
- deixar particularidades de systemd, Windows Service e launchd em adaptadores de instalacao, nao no dominio.

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

### 4.4.1 Experiencia desejada do shell

A sessao do shell deve se comportar como um ambiente vivo do SparkEdge.

Exemplo de fluxo:

```text
$ sparkedge shell
SparkEdge / local-edge > login
SparkEdge / local-edge > scripts list
SparkEdge / local-edge > scripts show 12
SparkEdge / local-edge > credentials create
SparkEdge / local-edge > servers discover 3
SparkEdge / local-edge > instances create
```

Regras desejadas:

- o prompt deve indicar o contexto atual;
- a sessao deve manter estado entre comandos;
- comandos comuns devem ficar acessiveis sem precisar decorar flags longas;
- o usuario deve poder alternar entre contexto global e contexto de entidade;
- o shell deve aceitar saida em tabela, texto ou JSON quando solicitado;
- o shell deve permitir ajuda contextual com `help` ou `?`;
- o shell deve permitir autocomplete dos comandos e das entidades mais usadas.

### 4.4.2 Navegacao contextual

O shell pode funcionar em niveis de contexto, como:

- ambiente global;
- usuario autenticado;
- script selecionado;
- servidor selecionado;
- instancia selecionada;
- device selecionado.

Isso permite fluxos como:

```text
SparkEdge > use script 12
SparkEdge / script:12 > show
SparkEdge / script:12 > edit
SparkEdge / script:12 > playground
```

Ou:

```text
SparkEdge > use instance 55
SparkEdge / instance:55 > logs
SparkEdge / instance:55 > run
SparkEdge / instance:55 > destinations
```

### 4.4.3 Comportamento minimo do shell

Mesmo no primeiro corte, o shell deveria suportar:

- login;
- listagem;
- visualizacao detalhada;
- criacao assistida;
- edicao simples;
- confirmacao antes de gravar;
- saida estruturada para automacao.

### 4.4.4 Quando o shell nao deve ser usado

Mesmo sendo especialmente util em dispositivos remotos e operacao com baixa latencia, o shell nao deve substituir totalmente os subcomandos.

Os subcomandos continuam importantes para:

- automacao em scripts;
- CI;
- chamadas remotas por SSH sem sessao interativa;
- uso rapido de uma unica acao;
- compatibilidade com integracoes futuras.

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

### Fase 0 - alinhamento e fundacao

- definir o nome final do binario principal;
- decidir se o shell sera o modo padrao ou um subcomando;
- mapear quais recursos sao obrigatorios no primeiro corte;
- definir o formato padrao de saida;
- definir como o CLI vai localizar o contexto local e a configuracao;
- validar como a sessao do shell sera persistida.

### Fase 1 - base do CLI

- criar entrypoint de comando;
- criar comandos base `version`, `health`, `config show`;
- padronizar saida JSON e texto;
- integrar com config local.

### Fase 1.1 - shell minimo funcional

- comando `sparkedge shell`;
- prompt com contexto local;
- comando `help`;
- comando `exit`;
- comando `clear`;
- comando `status`;
- listagem simples de entidades;
- retorno em texto e JSON.

### Fase 2 - autenticacao e usuarios

- login/logout;
- leitura de JWT local;
- status da sessao;
- CRUD minimo de usuarios.

### Fase 2.1 - comandos de leitura essenciais

- `auth whoami`;
- `users list`;
- `scripts list`;
- `servers list`;
- `instances list`;
- `credentials list`;
- `devices list`;
- `config validate`.

### Fase 3 - scripts

- listagem;
- install via ZIP;
- create/edit por arquivos;
- readme e playground no terminal;
- historico de versoes.

### Fase 3.1 - shell de scripts

- entrar em contexto de script;
- abrir README;
- ver arquivos do bundle;
- executar script em sandbox local;
- visualizar stdout e stderr;
- testar input manual antes de salvar;
- atualizar bundle sem recriar do zero.

### Fase 4 - credenciais, providers e servidores

- CRUD de credenciais;
- teste de conexao;
- descoberta de recursos;
- inspeccao dos schemas.

### Fase 4.1 - descoberta assistida

- listar providers disponiveis;
- testar credencial antes de salvar;
- exibir recursos e operacoes por provider;
- validar payloads e schemas via CLI.

### Fase 5 - instancias

- create/edit via CLI;
- escolha de script e device;
- mapeamento de payload e destino;
- execucao e logs.

### Fase 5.1 - shell de instancias

- entrar em contexto de instancia;
- ver configuracao resumida;
- editar trigger;
- editar destinos;
- revisar payload resolvido;
- disparar execucao manual;
- consultar logs da ultima execucao.

### Fase 6 - operacao e manutencao

- update assistido;
- rollback;
- backup;
- diagnostico;
- logs e health.

### Fase 6.1 - operacao remota assistida

- instalar release pelo terminal;
- validar checksum;
- preservar `config.yml`;
- preservar `sparkedge.db`;
- reiniciar com confirmacao;
- registrar historico da atualizacao.

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
- nao criar dependencia pesada que prejudique perfis low-resource, sem limitar desktop e servidor;
- nao fazer o CLI depender da WebUI;
- nao acoplar update assistido a internet o tempo todo.

## 11. Critérios de sucesso

O bloco de CLI pode ser considerado util quando:

- o mesmo fluxo funciona em desktop, servidor e edge compacto sem abrir a WebUI;
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

## 14. Backlog tecnico sugerido

### Marco A - navegação e identidade

- `sparkedge shell`;
- `sparkedge auth login`;
- `sparkedge auth whoami`;
- `sparkedge version`;
- `sparkedge health`.

### Marco B - leitura operacional

- `scripts list`;
- `servers list`;
- `instances list`;
- `devices list`;
- `credentials list`;
- `config show`.

### Marco C - escrita segura

- `users create`;
- `scripts create`;
- `scripts install`;
- `credentials create`;
- `servers create`;
- `instances create`.

### Marco D - contexto interativo

- trocar contexto dentro do shell;
- `use script <id>`;
- `use instance <id>`;
- `use server <id>`;
- `use device <id>`;
- `back`;
- `home`.

### Marco E - depuracao e suporte

- `scripts run`;
- `instances run`;
- `instances logs`;
- `instances history`;
- `servers discover`;
- `servers resources`.

### Marco F - operacao e recuperacao

- `update check`;
- `update apply`;
- `update rollback`;
- `backup create`;
- `config validate`.

## 15. Decisoes que ainda precisam fechar

- definir o nome final do binario principal;
- decidir se o shell sera o modo padrao ou um subcomando;
- mapear quais recursos sao obrigatorios no primeiro corte;
- definir o formato padrao de saida;
- definir como o CLI vai localizar o contexto local e a configuracao;
- validar como a sessao do shell sera persistida.

## 16. MVP implementavel

O primeiro corte deve resolver a operacao de manutencao em qualquer plataforma suportada, sem exigir a WebUI. Ele nao precisa reproduzir todos os formularios visuais, mas deve permitir identificar o estado do Edge, autenticar, consultar os recursos principais e executar as acoes mais comuns com seguranca.

### 16.1 Comandos obrigatorios do MVP

```text
sparkedge version
sparkedge health
sparkedge shell

sparkedge auth login
sparkedge auth logout
sparkedge auth whoami

sparkedge users list
sparkedge users create

sparkedge scripts list
sparkedge scripts show <id>
sparkedge scripts install <bundle.zip>
sparkedge scripts run <id>
sparkedge scripts playground <id>

sparkedge credentials list
sparkedge credentials create
sparkedge credentials test <id>

sparkedge servers list
sparkedge servers show <id>
sparkedge servers discover <id>

sparkedge instances list
sparkedge instances show <id>
sparkedge instances run <id>
sparkedge instances logs <id>

sparkedge config show
sparkedge config validate
```

Os comandos de criacao assistida de servidores e instancias entram logo depois da validacao desse corte, porque exigem mais perguntas interativas e maior cobertura dos schemas.

### 16.2 Contrato de saida

Todo comando deve poder ser executado em dois formatos:

```bash
sparkedge scripts list
sparkedge scripts list --output json
```

Regras do contrato:

- texto legivel e o formato padrao para uso humano;
- JSON deve ser estavel e adequado para scripts de shell;
- erros devem sair em `stderr` e retornar codigo diferente de zero;
- o JSON de erro deve conter `code`, `message` e, quando aplicavel, `details`;
- segredos nunca aparecem em listagens, logs ou mensagens de erro;
- `--quiet` deve permitir uso em automacoes sem mensagens decorativas;
- `--output json` nao pode depender de cores, largura do terminal ou idioma.

### 16.3 Sessao e armazenamento local

O shell deve manter apenas estado de sessao nao sensivel em memoria:

```text
SparkEdge / local > use script <id>
SparkEdge / local script:<id> > show
SparkEdge / local script:<id> > run --input input.json
```

Regras propostas:

- o JWT fica em armazenamento local com permissoes restritas, nunca no historico do shell;
- o caminho deve ser configuravel e seguir o diretorio de dados do Edge;
- `auth logout` remove a sessao local;
- o contexto selecionado nao deve alterar dados automaticamente;
- comandos destrutivos sempre exigem `--yes` ou confirmacao explicita;
- a sessao nao deve persistir senha, api key ou credential completa em texto puro.

### 16.4 Fluxo de criacao assistida

Para recursos complexos, o shell deve oferecer um fluxo guiado curto:

```text
sparkedge servers create --interactive
sparkedge instances create --interactive
```

O fluxo deve:

- apresentar somente perguntas relevantes ao provider ou ao script escolhido;
- permitir confirmar o resumo antes de gravar;
- aceitar `--from file.json` para repetir configuracoes;
- permitir cancelar sem deixar registros parciais;
- validar a configuracao antes da persistencia;
- informar qual campo falhou e mostrar um exemplo valido.

## 17. Criterios de aceite do primeiro corte

O MVP sera considerado operacional quando for possivel, apenas pelo terminal:

1. iniciar o SparkEdge em Windows, Linux, macOS e um alvo Linux ARM de baixo recurso sem WebUI;
2. verificar versao, saude, configuracao e banco local;
3. criar ou autenticar o primeiro usuario;
4. instalar um bundle de script e executar um teste com entrada JSON;
5. cadastrar uma credencial sem exibir o segredo depois da gravacao;
6. testar a conexao de um provider e listar seus recursos;
7. consultar uma instancia e disparar uma execucao manual;
8. obter stdout, stderr, exit code e logs da execucao;
9. repetir os mesmos fluxos com `--output json` em um script Bash;
10. executar os mesmos comandos sem conectividade com o SparkCloud.

## 18. Primeiras tarefas de implementacao

A implementacao deve seguir esta ordem:

1. extrair uma camada de servicos reutilizavel pelos controllers HTTP e pelo CLI;
2. criar o entrypoint Go e o comando `sparkedge version`;
3. integrar carregamento de `config.yml`, banco SQLite e logging;
4. implementar `health`, `config show` e `config validate`;
5. implementar `auth login`, `logout` e `whoami` usando o JWT existente;
6. implementar o renderer comum de texto, JSON e erros;
7. adicionar `shell`, `help`, `status`, `use`, `clear` e `exit`;
8. adicionar as listagens de scripts, servidores, credenciais e instancias;
9. adicionar instalacao e execucao de scripts;
10. validar no Windows, Linux amd64/arm64, macOS amd64/arm64 e Linux armv7 antes de iniciar os fluxos assistidos.

## 19. Fora do MVP inicial

Os itens abaixo permanecem no roadmap, mas nao devem bloquear a primeira entrega:

- TUI com paineis e navegacao completa;
- autocomplete semantico de todos os campos de payload;
- editor de codigo embutido no terminal;
- sincronizacao bidirecional de configuracao com o SparkCloud;
- execucao de workflows complexos pelo shell;
- instalacao automatica do binario por pacote do sistema.

## 20. Matriz de suporte multiplataforma

| Perfil | Sistema e arquitetura | Uso esperado | Validacao minima |
|---|---|---|---|
| Desktop Windows | Windows amd64 e arm64 | administracao local, desenvolvimento e operacao assistida | executavel, `config.yml`, SQLite, shell e WebUI |
| Desktop macOS | macOS amd64 e arm64 | administracao local e desenvolvimento | permissoes de arquivos, shell, SQLite e WebUI |
| Servidor Linux | Linux amd64 e arm64 | operacao continua, SSH e servico | sinais, logs, SQLite, update e restart |
| Edge ARM | Linux armv7 e armv6 | coleta remota e baixa conectividade | memoria, tempo de inicializacao, runner, shell e rollback |
| Automacao | qualquer alvo com shell nao interativo | scripts, CI/CD e diagnostico remoto | `--output json`, exit codes e ausencia de TTY |

A matriz nao significa que todos os alvos terao o mesmo mecanismo de instalacao. O binario e o comportamento do dominio devem ser comuns; o empacotamento e a integracao com o sistema operacional podem ser especificos.

### 20.1 Contratos que devem ser identicos

- nomes e argumentos dos subcomandos;
- formato JSON e codigos de erro;
- caminhos relativos resolvidos a partir de `data-dir` e `config.yml`;
- comportamento de autenticacao JWT;
- instalacao, execucao e historico de scripts;
- consulta de servidores, recursos, operacoes e instancias;
- logs de execucao e diagnosticos;
- compatibilidade offline com o banco SQLite local.

### 20.2 Adaptadores por plataforma

Somente as bordas devem variar:

- descoberta de diretorios padrao de configuracao e dados;
- leitura de sinais de encerramento;
- registro como Windows Service, systemd ou launchd;
- abertura opcional do navegador;
- permissao de arquivos e armazenamento de credenciais;
- formato do pacote de release e instalador.

### 20.3 Regras de distribuicao

Cada release deve publicar um pacote por combinacao suportada de sistema e arquitetura, com:

- executavel do SparkEdge;
- `webui/dist` quando a distribuicao incluir a WebUI;
- `config.yml` inicial ou exemplo seguro;
- documentacao de operacao;
- checksum do pacote;
- instrucoes de inicio, parada e rollback;
- identificacao explicita de OS, arquitetura e versao.

O pacote nao deve presumir que o usuario esta em um Raspberry, nem que possui Bash, PowerShell, systemd ou acesso ao SparkCloud. Quando uma dependencia externa for necessaria, o comando deve indicar a ausencia de forma clara e continuar oferecendo os fluxos locais possiveis.

## 21. Validacao por camadas

Para evitar que uma correcao para um alvo quebre outro, a verificacao deve ser dividida em camadas:

1. testes de dominio e CLI sem dependencia de sistema operacional;
2. testes de integracao com SQLite, filesystem, runner e configuracao;
3. testes de contrato dos comandos em texto, JSON e modo sem TTY;
4. build cruzado para cada combinacao publicada;
5. smoke test em Windows, Linux, macOS e pelo menos um alvo ARM;
6. teste de instalacao, restart, upgrade e rollback;
7. teste de degradacao sem SparkCloud e sem WebUI.

O Raspberry Pi 3 permanece importante como teste de estresse e compatibilidade armv7, mas a aprovacao da release depende da matriz completa, nao apenas desse dispositivo.
