# Используйте официальный образ Golang для сборки приложения
FROM golang:1.21 AS builder

# Установка переменных окружения
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Установка рабочей директории внутри контейнера
WORKDIR /app

# Копирование зависимостей проекта и загрузка их
COPY . .
#COPY microservice2/* .
RUN go mod download

# Копирование всех файлов вашего приложения
#COPY microservice2/*.go .

# Сборка приложения
RUN go build -o app .

# Отдельный этап для создания минимального образа
FROM alpine:latest

# Установка необходимых зависимостей (если они есть)
# RUN apk add --no-cache some-dependencies

# Копирование скомпилированного бинарного файла из предыдущего этапа
COPY --from=builder /app/app /app/app

# Установка рабочей директории внутри контейнера
WORKDIR /app

# Запуск вашего приложения
CMD ["./app"]