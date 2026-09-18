# ==========================================
# ETAPA 1: CONSTRUCTOR (Builder)
# ==========================================
# Usamos una imagen que tenga las herramientas de C (gcc, musl-dev) necesarias para compilar sqlite3
FROM golang:1.24-alpine AS builder

# Instalamos las dependencias de C necesarias para go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copiamos primero las dependencias para aprovechar la caché de Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código
COPY . .

# ⚠️ CORRECCIÓN CLAVE: CGO_ENABLED=1 y agregamos -tags sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -tags sqlite3 -o main ./cmd/api

# ==========================================
# ETAPA 2: IMAGEN FINAL (Runtime)
# ==========================================
FROM alpine:latest  

# Instalamos certificados SSL (necesarios para hacer peticiones HTTPS si las hubiera)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiamos el binario compilado desde la etapa builder
COPY --from=builder /app/main .

# Copiamos el archivo .env (opcional, pero útil si no usas variables de docker-compose)
COPY --from=builder /app/.env .

# ⚠️ CORRECCIÓN CLAVE: Creamos el directorio donde se guardará la BD SQLite
RUN mkdir -p /app/data

# El puerto interno real de tu aplicación es 3000 (según tu docker-compose.yml)
EXPOSE 3000

# Ejecutamos la aplicación
CMD ["./main"]
