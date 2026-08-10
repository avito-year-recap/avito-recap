
# syntax=docker/dockerfile:1

# Build the React application first.
# The resulting dist/ directory is copied into the final Go image,
# so production needs only one web service.

FROM node:24-alpine AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build


# Build the Go API.

FROM golang:1.25-alpine AS backend-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /out/api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /out/eventgen \
    ./cmd/eventgen


# Small runtime image: one process serves both React and the Go API.

FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=backend-builder /out/api /usr/local/bin/api
COPY --from=backend-builder /src/seeds ./seeds
COPY --from=frontend-builder /src/frontend/dist ./web

# Render provides PORT automatically.
# The Go config reads it when API_ADDRESS and HTTP_ADDR are not set.

ENV STORAGE_BACKEND=memory \
    PROFILES_PATH=/app/seeds/profiles.json \
    SCENARIOS_PATH=/app/seeds/scenarios.json \
    STATIC_DIR=/app/web \
    FRONTEND_DIR=/app/web

USER app

EXPOSE 8080

CMD ["/usr/local/bin/api"]


# eventgen: opt-in Kafka event generator (see docker-compose.yml's
# "events-gen" profile). Not part of the default `api` image/target above.

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

