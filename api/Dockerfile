# Etapa de Build
FROM golang:1.21rc3-alpine3.18 AS builder

ENV GOPATH /go

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de dependências go.mod e go.sum primeiro para melhorar a cache do Docker
COPY go.mod go.sum ./

# REMOVA SE ESITVER EM PRODUÇÃO - Copia o arquivo .env para o diretório de trabalho /app
COPY .env ./ 

# Baixa as dependências Go (não irá baixar novamente se não houver alterações no go.mod/go.sum)
RUN go mod download

# Copia o restante do código fonte para dentro do container
COPY . .

# Compila o binário do aplicativo Go
RUN go build -o /app/main .

# Etapa de Execução
FROM alpine:3.18 AS binary

# Definir o diretório de trabalho para a execução
WORKDIR /app

# Copia o binário compilado da etapa de Build para a imagem de execução
COPY --from=builder /app/main /app/main
#REMOVA SE ESITVER EM PRODUÇÃO
COPY .env ./

# Copiar a pasta 'static' para dentro do container
COPY static /app/static

# Exponha a porta que o aplicativo irá utilizar
ARG APP_PORT=8080
EXPOSE ${APP_PORT}

# Define o comando que será executado ao iniciar o container
CMD ["/app/main"]
