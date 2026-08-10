# Atualizacao assistida do SparkEdge

Este documento descreve a implementacao da estrategia de atualizacao assistida para o SparkEdge Go, usando o GitHub como fonte oficial de distribuicao.

## Objetivo

Entregar um fluxo seguro em que o usuario:

- verifica se existe nova versao;
- visualiza a release compativel com sua plataforma;
- decide quando baixar e aplicar;
- reinicia conscientemente o servico.

O foco inicial e reduzir risco operacional. Por isso, a primeira entrega nao substitui binarios automaticamente.

## Fases

### Fase 1 - Descoberta e checagem segura

Entregas:

- leitura da versao local e do alvo atual via `version.txt`;
- identificacao da plataforma atual no formato dos assets de release;
- configuracao de update em `config.yml`;
- consulta ao GitHub Releases;
- comparacao entre versao local e ultima release compativel;
- endpoint HTTP para a WebUI consultar update;
- tela inicial de "Atualizacao" na WebUI.

Resultado:

- o sistema passa a informar se existe atualizacao disponivel;
- o usuario ve versao atual, versao remota e asset compativel;
- nenhuma alteracao e aplicada ainda no disco.

### Fase 2 - Manifesto e checksums

Entregas:

- geracao de `manifest.json` por release;
- publicacao de `checksums.txt` ou equivalente;
- resolucao de assets preferencialmente pelo manifesto;
- validacao de integridade antes de qualquer download assistido.

Resultado:

- a camada de atualizacao fica menos dependente da estrutura textual dos assets;
- o sistema ganha uma verificacao de integridade mais forte.

### Fase 3 - Download assistido

Entregas:

- endpoint para baixar pacote da release correta;
- armazenamento do pacote em area temporaria de update;
- exibicao de progresso e estado na WebUI;
- persistencia do ultimo download e do ultimo erro.

Resultado:

- o usuario consegue preparar a atualizacao sem ainda aplicá-la.

### Fase 4 - Aplicacao assistida com backup

Entregas:

- criacao de backup da instalacao atual;
- extracao do pacote;
- validacao estrutural do conteudo esperado;
- troca controlada dos arquivos;
- reinicio assistido do processo;
- rollback se a troca falhar.

Resultado:

- atualizacao assistida completa, com foco em seguranca.

### Fase 5 - Aperfeicoamentos operacionais

Entregas:

- politicas por canal (`stable`, `beta`);
- agendamento opcional de checagem;
- historico de updates;
- integracao com service manager em Linux e Raspberry Pi.

## Configuracao proposta

O `config.yml` passa a poder conter:

```yaml
update:
  enabled: true
  provider: github
  repo: kelwSagashi/SparkEdgeGo
  channel: stable
  allow_prerelease: false
```

## Endpoints previstos

### Fase 1

- `GET /api/update/check`

### Fases posteriores

- `POST /api/update/download`
- `POST /api/update/apply`
- `POST /api/update/restart`

## Premissas de seguranca

- nenhuma atualizacao e aplicada sem acao explicita do usuario;
- a fase inicial apenas inspeciona metadata;
- aplicacao real de update exige checksum, backup e validacao do pacote;
- em Windows, a substituicao do executavel deve acontecer por processo auxiliar ou fluxo de reinicio controlado.
