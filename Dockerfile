FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 делает бинарник статическим, -o main - задает имя выходного файла
RUN CGO_ENABLED=0 go build -o main ./cmd/server/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app/

# Копируем готовый бинарник из этапа builder
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]