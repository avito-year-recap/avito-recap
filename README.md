# avito-recap

Backend for an Avito-style annual recap. The service turns a completed year of
seed activity into an immutable story: metrics, behavior, achievements,
personalized next action, and a privacy-safe public share card.

## Requirements

- Go 1.25.5 or newer
- Node.js 20 or newer (only for protobuf generation)

Docker is optional.

## Run locally

```powershell
go run ./cmd/api
```

The server listens on `http://localhost:8080` by default.

```text
GET  /health
GET  /avatars/{profile-code}.png
POST /recap.v1.RecapService/ListProfiles
POST /recap.v1.RecapService/GenerateRecap
POST /recap.v1.RecapService/GetRecap
POST /recap.v1.RecapService/GetShareCard
```

The RPC endpoints support Connect, gRPC, and gRPC-Web. Connect JSON example:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/recap.v1.RecapService/GenerateRecap" `
  -ContentType "application/json" `
  -Body '{"profileId":"26a3f4e0-1ae7-5b46-b2b6-2ae9fc180ba2","year":2025}'
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
| `API_ADDRESS` | `:8080` |
| `PROFILES_PATH` | `seeds/profiles.json` |
| `SCENARIOS_PATH` | `seeds/scenarios.json` |
| `STATIC_DIR` | `static` |
| `CORS_ALLOWED_ORIGINS` | localhost ports 3000 and 5173 |
| `SHUTDOWN_TIMEOUT` | `10s` |

## Tests

```powershell
go test ./...
```

Golden recap examples for frontend mocks are available under
`testdata/golden/`.

## Docker

```powershell
docker compose up --build
```