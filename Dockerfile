# Собираем докер по кусочкам - самое верхнее - часть с наименьшим изменением.
# Помогает оптимизировать кэщ докера

FROM golang:1.25-alpine AS backend-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

# Импортируем только то, что использует докер, чтоб не вызывать пересборку билда на каждое изменения документации
COPY cmd ./cmd
COPY internal ./internal
COPY gen ./gen

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /out/api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /out/eventgen \
    ./cmd/eventgen

COPY seeds ./seeds


# Упрощенная версия ОС под запуск приложения

FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=backend-builder /out/api /usr/local/bin/api
COPY --from=backend-builder /src/seeds ./seeds

ENV PROFILES_PATH=/app/seeds/profiles.json \
    SCENARIOS_PATH=/app/seeds/scenarios.json

USER app

EXPOSE 8080

CMD ["/usr/local/bin/api"]


# Опциональная часть - режим работы реального общего сервиса

FROM alpine:3.20 AS eventgen

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=backend-builder /out/eventgen /usr/local/bin/eventgen
COPY --from=backend-builder /src/seeds ./seeds

ENV PROFILES_PATH=/app/seeds/profiles.json \
    SCENARIOS_PATH=/app/seeds/scenarios.json

USER app

CMD ["/usr/local/bin/eventgen"]

