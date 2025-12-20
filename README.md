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
- SendGrid API key (opcional, para emails)
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
- **Email**: SendGrid
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
