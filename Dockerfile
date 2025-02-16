# Этап сборки
FROM golang:1.23.5-alpine3.21 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем только go.mod, чтобы избежать повторной загрузки зависимостей
COPY go.mod ./

# Проверяем наличие go.sum и создаем его, если отсутствует
RUN [ -f go.sum ] || touch go.sum

# Устанавливаем зависимости
RUN go mod download

# Копируем весь исходный код в контейнер
COPY . .

# Передаем аргументы сборки для версии, коммита и времени сборки
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_TIME=unknown

# Собираем приложение с флагами для внедрения версии
RUN go build -ldflags="-X goindex/handlers.Version=$VERSION -X giondex/handlers.GitCommit=$GIT_COMMIT -X goindex/handlers.BuildTime=$BUILD_TIME" -o app .

# Финальный образ
FROM alpine:3.21.2

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем скомпилированное приложение
COPY --from=builder /app/app .

# Экспонируем порт для HTTP
EXPOSE 8080

# Устанавливаем команду запуска
CMD ["./app"]