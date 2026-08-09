FROM golang:1.25.5-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /src/seeds ./seeds
COPY --from=builder /src/static ./static
USER app

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
