<div align="center">
  <img src="https://cdn.jsdelivr.net/gh/devicons/devicon/icons/go/go-original.svg" height="120" alt="go logo"  />

  # API de Autenticação Segura 
  
  *Um serviço robusto de autenticação e gestão de usuários, focado em segurança de alto nível para aplicações.*
</div>

---

## 📖 Sobre o Projeto

Este repositório contém uma API de autenticação robusta e segura construída com Go. O projeto oferece um sistema completo de gerenciamento de usuários, incluindo registro, login, verificação de e-mail, redefinição de senha e gerenciamento de sessão via JWT. A aplicação é totalmente containerizada com Docker e sua infraestrutura é provisionada na Google Cloud Platform (GCP) usando Terraform, com um pipeline de CI/CD automatizado através do GitHub Actions.

## ✨ Features

-   **Gerenciamento de Usuários:** Cadastro, login e logout.
-   **Segurança de Senha:** Hashing de senhas utilizando o algoritmo **Argon2id**.
-   **Autenticação baseada em JWT:** Uso de `access_token` e `refresh_token` para sessões seguras.
-   **Gerenciamento de Cookies:** Tokens armazenados em cookies `HttpOnly`, `Secure` e `SameSite=Lax` para maior segurança contra ataques XSS.
-   **Verificação de E-mail:** Fluxo de confirmação de conta com envio de código de verificação por e-mail (integrado com MailerSend).
-   **Recuperação de Senha:** Funcionalidade de "Esqueci minha senha" com envio de código para redefinição.
-   **Middleware de Proteção:** Rotas protegidas que exigem autenticação válida.
-   **Containerização:** Aplicação e banco de dados gerenciados via Docker e Docker Compose.
-   **Infraestrutura como Código (IaC):** Configuração de VM e firewall na GCP gerenciada com Terraform.
-   **CI/CD:** Pipeline automatizado com GitHub Actions para testes e deploy contínuos na `master`.

## 🛠️ Tecnologias Utilizadas

-   **Backend:** Go (Golang)
-   **Banco de Dados:** PostgreSQL
-   **Containerização:** Docker, Docker Compose
-   **Reverse Proxy:** Caddy
-   **Infraestrutura:** Terraform, Google Cloud Platform (GCP)
-   **CI/CD:** GitHub Actions
-   **Serviço de E-mail:** MailerSend

## 🚀 Como Rodar Localmente

### Pré-requisitos

-   [Docker](https://www.docker.com/get-started)
-   [Docker Compose](https://docs.docker.com/compose/install/)

### Instalação

1.  **Clone o repositório:**
    ```bash
    git clone https://github.com/jeanpdr/api_auth_go.git
    cd api_auth_go
    ```

2.  **Crie o arquivo de ambiente:**
    Crie um arquivo chamado `.env` na raiz do projeto, copiando o conteúdo do exemplo abaixo.

3.  **Inicie os containers:**
    Execute o comando a seguir para construir e iniciar a API, o banco de dados e o reverse proxy.
    ```bash
    docker-compose up --build -d
    ```
    -   A API estará disponível em `http://localhost:8080`.
    -   O banco de dados estará acessível na porta `5432` da sua máquina local.

## ⚙️ Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto com as seguintes variáveis:

```env
# Configuração do Banco de Dados PostgreSQL
DB_USER=seu_usuario_db
DB_PASSWORD=sua_senha_db
DB_NAME=seu_nome_db
# DB_HOST já é definido no docker-compose.yaml para o serviço da API

# Chave secreta para assinatura dos tokens JWT
JWT_SECRET=sua_chave_secreta_super_longa_e_segura

# Configuração do MailerSend (Serviço de E-mail)
MAILERSEND_API_KEY=sua_api_key_do_mailersend
MAILERSEND_FROM_EMAIL=seu_email_verificado_no_mailersend
MAILERSEND_FROM_NAME="Nome do Seu App"
MAILERSEND_TEMPLATE_ID=seu_template_id_do_mailersend
```

## 🔌 Endpoints da API

| Método | Rota                       | Descrição                                                              | Autenticação |
| :----- | :------------------------- | :--------------------------------------------------------------------- | :----------: |
| `GET`  | `/health`                  | Verifica o status da API.                                              |      Não       |
| `POST` | `/register`                | Registra um novo usuário.                                              |      Não       |
| `POST` | `/login`                   | Autentica um usuário e retorna cookies de sessão.                      |      Não       |
| `POST` | `/verify`                  | Verifica o e-mail do usuário com o código recebido.                    |      Não       |
| `POST` | `/verify/resend`           | Reenvia um novo código de verificação para o e-mail.                   |      Não       |
| `POST` | `/forgot-password`         | Inicia o fluxo de recuperação de senha.                                |      Não       |
| `POST` | `/reset-password`          | Altera a senha usando o código de recuperação.                         |      Não       |
| `POST` | `/refresh`                 | Gera um novo `access_token` usando o `refresh_token`.                  |      Sim       |
| `POST` | `/logout`                  | Invalida os tokens de sessão e limpa os cookies.                       |      Sim       |
| `GET`  | `/dashboard`               | Exemplo de rota protegida.                                             |      Sim       |
| `GET`  | `/me`                      | Retorna o status de autenticação do usuário.                           |      Sim       |

Link da documentação da API -> https://docs-api-go.jeanpreis.com.br/

## ☁️ Infraestrutura e Deploy (CI/CD)

### Terraform

A pasta `terraform` contém os arquivos para provisionar a infraestrutura na **Google Cloud Platform (GCP)**. O script cria:
-   Uma instância de VM `e2-micro` (dentro do Free Tier da GCP).
-   Uma regra de firewall para permitir tráfego nas portas `80`, `443`, `8080` e `22`.

### GitHub Actions

O fluxo de trabalho `ci.yml` automatiza os seguintes processos:

1.  **Test and Build**:
    -   Executado em cada `push` ou `pull_request` para a branch `master`.
    -   Inicia um serviço do PostgreSQL.
    -   Executa todos os testes unitários do projeto (`go test`).

2.  **Deploy**:
    -   Executado somente em `push` para a branch `master`.
    -   Conecta-se via SSH à VM na GCP.
    -   Executa `git pull` para obter a versão mais recente do código.
    -   Executa `docker-compose up --build -d` para reconstruir e reiniciar os serviços.
    -   Limpa imagens antigas do Docker para economizar espaço em disco.

Para o deploy funcionar, os seguintes secrets precisam ser configurados no repositório do GitHub:
-   `GCP_HOST`: O IP público da VM.
-   `GCP_USERNAME`: O nome de usuário para o acesso SSH.
-   `GCP_SSH_KEY`: A chave SSH privada para autenticação.


