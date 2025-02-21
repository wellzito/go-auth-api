# Meu Projeto Go

Este projeto é uma API escrita em Go, que utiliza PostgreSQL como banco de dados de salvamento, Redis como banco de dados de cache 
e é gerenciada com Docker Compose.

## 🚀 Requisitos

Antes de começar, certifique-se de ter instalado em sua máquina:
- [Go](https://go.dev/dl/) (versão 1.21 ou superior)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

## 📦 Configuração do Banco de Dados

Este projeto utiliza o PostgreSQL como banco de dados. Para configurar e rodar o banco de dados corretamente:

1. **Criação do banco de dados**

   O banco de dados será gerado automaticamente ao iniciar o serviço via Docker Compose.

2. **Rodar o Docker Compose**

   Para subir o banco de dados, utilize o seguinte comando:

   ```sh
   docker-compose up --build
   ```

   **OU**, se houver problemas com esse comando, tente:

   ```sh
   docker compose up -d
   ```

   O argumento `-d` faz com que os serviços rodem em segundo plano.

3. **Verificar se o banco está rodando**

   ```sh
   docker ps
   ```
   Esse comando deve listar um contêiner do PostgreSQL em execução.

## 🛠 Rodando a API

Após iniciar o banco de dados, execute a aplicação com Docker:

```sh
docker build -t meu-projeto-go .
docker run -p 8080:8080 meu-projeto-go
```

Caso queira rodar a aplicação sem Docker:

```sh
go run main.go
```

## 🔧 Variáveis de Ambiente

Este projeto utiliza um arquivo `.env` para configurar as credenciais do banco de dados e outras configurações essenciais. Certifique-se de criar um arquivo `.env` na raiz do projeto com o seguinte conteúdo:

```env
# Porta onde o servidor da aplicação escuta
APP_PORT=""

# Chave secreta para assinar o token
SECRET_KEY=""

# Banco de dados PostgreSQL
DB_HOST=""
DB_NOME=""
DB_PORT=""
DB_USUARIO=""
DB_SENHA=""

# REDIS
REDIS_URL=""
```

O formato da URL de conexão com o PostgreSQL deve ser algo como:

```env
DATABASE_URL=postgres://$DB_USUARIO:$DB_SENHA@localhost:5432/$DB_NOME?sslmode=disable
```

Se estiver em produção, será:

```env
DATABASE_URL=postgres://$DB_USUARIO:$DB_SENHA@localhost:5432/$DB_NOME?sslmode=require
```

## 📜 Comandos Úteis

- **Parar os serviços do Docker Compose:**
  ```sh
  docker compose down
  ```

- **Reiniciar os serviços:**
  ```sh
  docker compose restart
  ```

- **Acessar o banco via CLI:**
  ```sh
  docker exec -it partento psql -U usuario -d nome_do_banco
  ```

- **Se precisar parar e limpar tudo antes de tentar de novo:**
  ```sh
  docker compose down --volumes --remove-orphans
  ```
  
- **Reconstrua e suba tudo de novo:**
  ```sh
  docker compose up --build
  ```
  
- **Se quiser rodar em background (modo daemon), use:**
  ```sh
  docker compose up -d --build
  ``` 

- **Recrie a Imagem:**
  ```sh
  docker-compose build --no-cache
  docker-compose up
  ```   
## 📜 Importantes

- **Redis:**
  ```sh
  Este projeto usa o Redis no Login para evitar muitas chamadas ou ataques na API
  o User é bloqueado por algum tempo e depois que o tempo expirar, poderá fazer login novamente
  go get github.com/go-redis/redis/v8
  ``` 
  
  ```sh
  O token de login também é salvo em cache para que recupe rapidamente, sem ter que chamar
  o PostgreSQL constantemente. O cache tem duração de 15 minutos determinada no codigo, em login.go 'timeWindow'
  ``` 
  

## ❓ Possíveis Erros

### `unable to prepare context: path "./api" not found`

Se encontrar esse erro ao rodar `docker compose up`, pode ser necessário verificar o caminho correto no `docker-compose.yml`. Certifique-se de que a estrutura de diretórios está correta e de que o contexto do `Dockerfile` está apontando para o local correto.

Caso o problema persista, tente rodar o comando dentro da pasta `api`:

```sh
cd api
docker compose up -d
```

## Produção

Se for fazer deploy em produção, remova as seguintes linhas:

- **docker-compose.yml:**

Linhas 5/6 e 45/46
	```sh
    env_file: # Remova essa linha e a linha abaixo se estiver em produção
      - .env  # Carrega o arquivo .env com as variáveis de ambiente
	```
	
- **Dockerfile:**

Linhas 12/13 e 32/33
	```sh
    COPY .env ./ 
	```

- **main.go:**

Linhas 18 até 22
	```sh
    // Carrega as variáveis do .env remover se tiver em produção
	erro := godotenv.Load()
	if erro != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}
	```

- **config.go:**

Linhas 34 até 39
	```sh
    //Descomente para usar local e comente para usar em prod
	// Carregar as variáveis de ambiente primeiro
	//Remover função abaixo se tiver em produção
	if erro = godotenv.Load(); erro != nil {
		log.Fatal(erro)
	}
	```	