# Etapa 1: Build (Compilação)
# Usamos uma imagem leve do Go baseada em Alpine Linux
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copia os arquivos de dependência primeiro (para aproveitar o cache do Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte do projeto
COPY . .

# Compila o binário.
# CGO_ENABLED=0 garante um binário estático puro (sem dependências de C)
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Etapa 2: Runtime (Execução)
# Usamos uma imagem vazia e mínima do Alpine para rodar
FROM alpine:latest

WORKDIR /app

# Instala certificados de segurança (necessário para fazer chamadas HTTPS para a Funceme)
RUN apk --no-cache add ca-certificates

# Copia o binário gerado na etapa anterior
COPY --from=builder /app/main .

# Copia a pasta estática (imagens) para dentro do container
COPY static ./static

# Expõe a porta que a aplicação usa
EXPOSE 8000

# Comando para iniciar a aplicação
CMD ["./main"]
