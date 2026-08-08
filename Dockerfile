FROM golang:1.25.5-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/recap-api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache wget \
	&& addgroup -S recap \
	&& adduser -S -G recap recap
WORKDIR /app
COPY --from=build /out/recap-api /app/recap-api
COPY seeds /app/seeds
COPY static /app/static

USER recap
EXPOSE 8080
ENTRYPOINT ["/app/recap-api"]
