# avito-recap

Backend for an Avito-style annual recap. The service turns a completed year of
seed activity into an immutable story: metrics, behavior, achievements,
personalized next action, and a privacy-safe public share card.


## Deploy to Render (jury demo)

The repository is prepared for a **single Render Web Service**. In production
the Docker image builds React, builds Go, and the Go process serves both the SPA
and the Connect API. The demo uses the in-memory seed storage, so a separate
ClickHouse service is not required on Render.

Architecture:

```text
browser -> https://<service>.onrender.com
            |-- /, /recap/*, /share/* -> React SPA
            |-- /api/*                -> Go Connect API
            |-- /health               -> health check
```

Deployment steps:

1. Push this repository to GitHub.
2. In Render choose **New -> Blueprint** and select the repository.
3. Render reads `render.yaml` and builds the root `Dockerfile`.
4. Wait until the service becomes `Live`.
5. Open the generated `https://<service>.onrender.com` URL and run the demo flow.

You can also create a **Web Service** manually and select the `Docker` runtime.
Use the repository root as the Docker context and `/health` as the health-check
path. Do not set `PORT` yourself; Render provides it and the application reads it
automatically.

The Render demo intentionally uses `STORAGE_BACKEND=memory`: all 17 seed
profiles work, generated recaps live for the lifetime of the process, and no
external database is required. The existing `docker compose` development setup
still uses ClickHouse.

On the free Render plan the service can spin down after inactivity. After a
restart the seed catalogue is loaded again automatically, but previously
generated recap/share IDs are intentionally not persisted in memory mode.

## Requirements

- Go 1.25.5 or newer (the Render Docker image tracks the latest Go 1.25 patch)
- Node.js 24 recommended for frontend tooling

## Run locally

Install frontend dependencies once after cloning/unpacking the repository:

```powershell
cd frontend
npm ci
cd ..
```

This also makes VS Code/TypeScript resolve `@connectrpc/connect` and `vite/client`.

```powershell
go run ./cmd/api
```

```powershell
docker compose up --build
```

The server listens on `http://localhost:8080` by default.

```text
GET  /health
GET  /avatars/{profile-code}.png
POST /recap.v1.RecapService/ListProfiles
POST /recap.v1.RecapService/GenerateRecap
POST /recap.v1.RecapService/GetRecap
POST /recap.v1.RecapService/GetPublicShare
```

The RPC endpoints support Connect, gRPC, and gRPC-Web. Connect JSON example:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/recap.v1.RecapService/GenerateRecap" `
  -ContentType "application/json" `
  -Body '{"profileCode":"active-buyer","year":2025}'
```

The demo catalogue contains 17 profiles from `seeds/profiles.json`. All bundled
scenarios use the completed year `2025`.

## Frontend integration

Generated TypeScript schemas and the service descriptor are committed at:

```text
gen/ts/recap/v1/recap_pb.ts
```

Install the client runtime:

```powershell
npm install @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-web
```

Create a typed browser client:

```typescript
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { RecapService } from "./gen/ts/recap/v1/recap_pb";

const transport = createConnectTransport({
  baseUrl: "http://localhost:8080",
});

export const recapClient = createClient(RecapService, transport);
```

Profile avatar URLs are relative to the API origin. Resolve `/avatars/...`
against the same `baseUrl` when the frontend runs on a different origin.

Default CORS origins are `http://localhost:3000` and
`http://localhost:5173`. Override them with a comma-separated
`CORS_ALLOWED_ORIGINS` value.

## Generate API code

The protobuf contract lives at `proto/recap/v1/recap.proto`.

```powershell
npx --yes @bufbuild/buf@1.72.0 lint
npx --yes @bufbuild/buf@1.72.0 generate
```

Generation produces:

- Go protobuf messages in `gen/go/recap/v1`
- Connect Go clients and handlers in `gen/go/recap/v1/recapv1connect`
- TypeScript schemas and service descriptors in `gen/ts/recap/v1`

## Configuration

| Variable | Default |
| --- | --- |
| `API_ADDRESS` | empty; falls back to `HTTP_ADDR`, then Render `PORT`, then `:8080` |
| `STORAGE_BACKEND` | `clickhouse` (`memory` in the Render Docker image) |
| `CLICKHOUSE_DSN` | `clickhouse://recap:recap@clickhouse:9000/recap` |
| `PROFILES_PATH` | `seeds/profiles.json` |
| `SCENARIOS_PATH` | `seeds/scenarios.json` |
| `STATIC_DIR` | `static` |
| `FRONTEND_DIR` | empty (set to `/app/web` in the Render Docker image) |
| `CORS_ALLOWED_ORIGINS` | localhost ports 3000 and 5173 |
| `SHUTDOWN_TIMEOUT` | `10s` |

## Tests

```powershell
go test ./...
```

For only the Render single-service integration tests:

```powershell
make test-render
```

Render-style integration coverage is in `internal/server/render_integration_test.go`.
It verifies the SPA/deep-link fallback, the complete recap flow through `/api`,
public sharing, static avatars, same-origin requests, and API 404 behavior.

ClickHouse-only integration tests remain behind the `integration` build tag and
require a running ClickHouse instance.

Golden recap examples for frontend mocks are available under
`testdata/golden/`.

## Docker

```powershell
docker compose up --build
```