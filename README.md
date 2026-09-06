# UspAvalia

Sistema de avaliação de disciplinas e professores da USP (Universidade de São Paulo).

Para mais informações, acesse [https://uspavalia.com/sobre](https://uspavalia.com/sobre).

## Quick Start com Docker

A forma mais simples de executar o UspAvalia:

```bash
docker-compose up -d
```

Acesse: http://localhost:8080

## Pré-requisitos

- **Docker** e **Docker Compose** (para deployment com containers)
- **Go 1.24.6+** (para desenvolvimento local)
- **SQLite** ou **MySQL** (banco de dados)

## Desenvolvimento Local

### 1. Configuração

Copie o arquivo de exemplo e configure:

```bash
cp .uspavalia.example.yaml .uspavalia.yaml
```

Edite `.uspavalia.yaml` com suas configurações:
- Chaves de segurança (secret_key, csrf_key, magic_link_hmac_key)
- Google OAuth credentials
- AWS SES credentials (opcional, para emails)
- hCaptcha keys (opcional, para proteção contra bots)

**Importante**: Gere chaves seguras para produção:
```bash
openssl rand -base64 32
```

### 2. Build

```bash
go build -o uspavalia .
```

### 3. Inicializar Banco de Dados

```bash
# Criar tabelas
./uspavalia migrate

# Importar unidades do Jupiter Web
./uspavalia fetch-units --store

# Importar disciplinas e professores
./uspavalia fetch-disciplines --store
```

### 4. Executar Servidor

```bash
./uspavalia serve
```

Acesse: http://localhost:8080

## Configuração

O aplicativo suporta dois métodos de configuração:


## Autenticação

O sistema suporta dois métodos de autenticação:

1. **Google OAuth**: Login social com conta Google
2. **Magic Link**: Login sem senha via link enviado por email

### Proteção contra Bots

- **hCaptcha**: Proteção em formulários de login, registro e contato
- Configure as chaves em `security.hcaptcha_site_key` e `security.hcaptcha_secret_key`

## Arquitetura

- **Linguagem**: Go 1.24.6+
- **Framework Web**: Gorilla Mux
- **ORM**: GORM (SQLite/MySQL)
- **Templates**: Go html/template
- **Autenticação**: OAuth 2.0 (Google), Magic Links (HMAC)
- **Email**: AWS SES
- **Segurança**: CSRF protection, rate limiting, hCaptcha

### Estrutura de Diretórios

```
.
├── cmd/                    # Comandos CLI (Cobra)
├── internal/
│   ├── config/            # Configuração (Viper)
│   ├── database/          # Database e migrations
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Models GORM
│   └── services/          # Serviços (email, etc)
├── pkg/                   # Pacotes públicos
│   ├── auth/              # Autenticação e crypto
│   └── utils/             # Utilidades
├── templates/             # Templates HTML
├── static/                # CSS, JS, imagens
└── matrusp/              # Dados do MatrUSP
```

## Testes

```bash
go test -v ./...
```

## Monitoramento

### Métricas Prometheus

Endpoint: `http://localhost:8080/metrics`

Métricas disponíveis:
- Duração de requisições HTTP
- Conexões ativas
- Total de requisições
- Status do pool de conexões do banco
- Total de usuários ativos
- Última execução do scraper do Júpiter Web (veja abaixo)

### Métricas do scraper

Cada execução de `fetch-disciplines --store` grava uma linha na tabela
`scrape_runs` (início, fim, status e contadores). O servidor lê a última linha e
expõe em `/metrics`, todas com o label `command` (hoje só `fetch-disciplines`):

| Métrica | Descrição |
|---|---|
| `uspavalia_scrape_last_run_status{status}` | 1 no status atual (`running`, `success`, `partial`, `failed`), 0 nos demais |
| `uspavalia_scrape_last_run_started_timestamp_seconds` | Início da última execução (Unix) |
| `uspavalia_scrape_last_run_finished_timestamp_seconds` | Fim da última execução (ausente enquanto roda) |
| `uspavalia_scrape_last_run_duration_seconds` | Duração da última execução |
| `uspavalia_scrape_last_success_timestamp_seconds` | Fim da última execução que gravou dados (`success` ou `partial`) |
| `uspavalia_scrape_last_run_pages_requested` | Páginas do Júpiter requisitadas |
| `uspavalia_scrape_last_run_pages{result}` | Páginas por resultado: `loaded`, `http_error`, `bad_status`, `parse_error` |
| `uspavalia_scrape_last_run_disciplines{stage}` | Disciplinas por etapa: `listed`, `processed`, `skipped`, `stored` |
| `uspavalia_scrape_last_run_units_found` | Unidades de ensino encontradas |
| `uspavalia_scrape_last_run_unit_list_errors` | Unidades cuja lista de disciplinas falhou |
| `uspavalia_scrape_last_run_offerings_stored` | Turmas criadas ou atualizadas |
| `uspavalia_scrape_last_run_professors_created` | Professores novos |
| `uspavalia_scrape_last_run_store_errors` | Erros de escrita no banco |
| `uspavalia_scrape_runs_total{status}` | Total acumulado de execuções por status |

Status: `failed` quando houve erro fatal ou nenhuma disciplina foi gravada;
`partial` quando houve qualquer erro de HTTP, parse ou banco; `success` caso
contrário. Uma execução que morreu no meio fica como `running` para sempre, o
que aparece como início antigo sem fim.

Exemplos de consultas para Grafana/alertas:

```promql
# Horas desde a última execução que gravou dados (alerta se > 8 dias = 192h)
(time() - uspavalia_scrape_last_success_timestamp_seconds) / 3600

# Última execução falhou
uspavalia_scrape_last_run_status{status="failed"} == 1

# Execução travada: começou há mais de 6h e ainda está "running"
uspavalia_scrape_last_run_status{status="running"} == 1
  and (time() - uspavalia_scrape_last_run_started_timestamp_seconds) > 6 * 3600

# Taxa de páginas com problema na última execução
sum(uspavalia_scrape_last_run_pages{result!="loaded"}) / uspavalia_scrape_last_run_pages_requested

# Execuções por status na última semana
increase(uspavalia_scrape_runs_total[7d])
```

### Compatibilidade

Este projeto foi reescrito de PHP para Go em 2024-2025. O código PHP original (2014) ainda está presente no repositório para referência, mas não é mais usado.

- URLs antigas (`/?p=disciplina&id=123`) são automaticamente redirecionadas
- Dados podem ser importados com `./uspavalia import-old-data`

## Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Contato

Para dúvidas ou sugestões, use o formulário de contato em [uspavalia.com/contato](https://uspavalia.com/contato).

## Notas de Segurança

- **NÃO** commite arquivos com credenciais (`.env`, `.uspavalia.yaml` com valores reais)
- **NÃO** use as chaves de exemplo em produção
- Gere chaves seguras para produção: `openssl rand -base64 32`
- Configure HTTPS em produção (via proxy reverso como Nginx/Traefik)
- Emails dos usuários são armazenados como hash SHA256, nunca em plaintext
