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
still uses ClickHouse. Demo event seeding there is restart-idempotent by profile/year
event count, so restarting only the API does not append the same event year again. If
you edit `seeds/scenarios.json`, reset the local demo volume before reseeding with
`docker compose down -v`; a mismatched existing event count is rejected instead of
silently inflating annual metrics.

The Render blueprint keeps `AI_NARRATIVE_PROVIDER=off`: the AI layer is now
local-only and expects an Ollama server reachable from the backend machine. The
published Render demo therefore uses the deterministic recap copy unless you
separately deploy an Ollama-compatible host and override the configuration.

On the free Render plan the service can spin down after inactivity. After a
restart the seed catalogue is loaded again automatically, but previously
generated recap/share IDs are intentionally not persisted in memory mode.

## Requirements

- Go 1.25.5 (the Docker builder is pinned to the same patch for reproducible builds)
- Node.js 24 recommended for frontend tooling
- Ollama (optional, only for local AI storytelling)

## Run locally

Install frontend dependencies once after cloning/unpacking the repository:

```powershell
cd frontend
npm ci
cd ..
```

This also makes VS Code/TypeScript resolve `@connectrpc/connect` and `vite/client`.

For local AI storytelling, install Ollama and pull the default model once:

```powershell
ollama pull qwen3:4b
```

When the Go API is started directly on the host, the default `auto` mode checks
`http://localhost:11434` and enables narrative generation only when Ollama is
running and the configured model exists:

```powershell
$env:AI_NARRATIVE_PROVIDER="auto"
go run ./cmd/api
```

To require Ollama instead of silently falling back at startup:

```powershell
$env:AI_NARRATIVE_PROVIDER="ollama"
go run ./cmd/api
```

For Docker Compose, keep Ollama running on the host and start the project normally:

```powershell
docker compose up --build
```

The API container reaches the host Ollama process through
`http://host.docker.internal:11434`; `extra_hosts: host-gateway` is included for
Linux Docker installations as well.

The server listens on `http://localhost:8080` by default.

```text
GET  /health
GET  /avatars/{profile-code}.png
GET  /api/explain?profile_code={code}&year={year}
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

### Explainability API

`GET /api/explain?profile_code=active-buyer&year=2025` returns a privacy-safe
decision trace for the generated recap. It shows:

- the ruleset version and digest;
- every behavior candidate, its priority and the checks that passed/failed;
- the selected achievements and their reasons;
- every next-action candidate and which rule won by priority.

The trace intentionally omits internal profile UUIDs and executable listing/dialog
target IDs. It is designed for rule debugging, support tooling and a jury demo of
why a concrete user received a concrete recap.

### AI Narrative Generator

The optional AI layer runs **after** metrics, behavior, achievements and next action
have already been selected by the deterministic rule engine. AI cannot change card
IDs/types, payloads, explanations, achievements, behavior or executable CTA targets.
It may replace only `description` for an explicit allow-list of personal recap cards:
`INTRO`, `YEAR_ACTIVITY`, `TOP_CATEGORY`, `ACTIVE_MONTH`, `BEHAVIOR`, `ACHIEVEMENT`,
`MISSED_OPPORTUNITY` and `NEXT_ACTION`. The public `SHARE` card is deliberately excluded
and always keeps deterministic privacy-reviewed copy.

```text
raw events -> metrics -> deterministic rules -> recap cards
                                             -> safe aggregate facts
                                             -> local Ollama /api/chat
                                             -> description overrides only
```

The project no longer requires a paid external LLM API key. The concrete adapter is
`internal/ai/ollama`, which talks to the local Ollama HTTP API. The default model is
`qwen3:4b`; pull it once with `ollama pull qwen3:4b`. You may select another locally
installed Ollama model through `OLLAMA_MODEL` without changing the recap engine.

The adapter uses Ollama structured outputs: it sends the expected JSON Schema in the
`format` field and also includes the schema in the user prompt. Streaming and model
thinking are disabled for this short copywriting request, so the backend receives one
JSON object in `message.content` and decodes it into `narrative.Story`.

Privacy boundary: Ollama receives only the recap year, aggregate counters/rates,
top-category label, active-month label, already-selected behavior, achievement
summaries, already-selected next action and the IDs of AI-editable cards only. The
`SHARE` card ID is not sent to the model. Profile UUIDs, share IDs, listing/dialog IDs,
raw events, message text and exact purchase
objects are not passed to the narrative provider. With a local Ollama server these
facts stay on the machine running Ollama rather than being sent to a paid cloud LLM.

Provider failures are **best effort** after startup: timeout, runtime model error or
malformed generation does not break recap generation; the original deterministic
descriptions remain in place. AI output is atomic: a non-empty response must contain
exactly one description for every editable card. Partial, duplicate, unknown or
`SHARE` overrides are rejected as a whole and logged through the same narrative error
callback as provider failures. Stored recap integrity permits only allowed
`Card.Description` values to differ from the deterministic card projection; IDs,
types, order, titles, explanations, share flags and payloads are revalidated on reads.

Provider modes:

```text
AI_NARRATIVE_PROVIDER=auto
  -> startup checks local Ollama and OLLAMA_MODEL
  -> available: AI enabled
  -> unavailable: deterministic copy, no startup failure

AI_NARRATIVE_PROVIDER=ollama
  -> local Ollama/model must be available at startup
  -> startup fails with a clear error if the model was not pulled

AI_NARRATIVE_PROVIDER=off
  -> AI completely disabled
```

Typical local settings:

```powershell
$env:AI_NARRATIVE_PROVIDER="ollama"
$env:OLLAMA_MODEL="qwen3:4b"
$env:OLLAMA_BASE_URL="http://localhost:11434"
$env:OLLAMA_KEEP_ALIVE="5m"
go run ./cmd/api
```

No `OPENAI_API_KEY` or other paid-provider key is used by this version.

### Concurrency controls

Recap generation has two explicit concurrency guards:

- **Per-key singleflight** in `application.Service.Generate`: concurrent requests with
  the same `profileID + year + rulesVersion + rulesDigest` share one expensive
  generation. Metrics, rule evaluation and the optional AI request run once, while
  followers wait for the same result. Each HTTP caller can cancel independently; the
  shared generation is canceled only after its last waiter leaves. This guard is
  process-local: the current ClickHouse adapter is single-writer and does not provide
  an atomic cross-replica uniqueness guarantee, so production multi-replica deployment
  would need a distributed lock or a dedicated OLTP uniqueness boundary.
- **AI semaphore** in `narrative.Limited`: local Ollama inference calls are capped by
  `AI_NARRATIVE_MAX_CONCURRENCY` (default `2`) per application process. Additional recap
  requests wait for a slot and respect request-context cancellation. This prevents a
  burst of recap requests from starting too many simultaneous model inferences and
  exhausting local CPU/GPU/RAM. If several backend replicas share one Ollama host, a
  host-wide concurrency limit would still need coordination outside each process.

These controls solve different problems: singleflight removes duplicate work for the
_same recap_, while the semaphore bounds AI calls across _different recaps_.

### Achievement rules v3.6

Seller achievements now use conversion percentages instead of only absolute sale
counts:

- **Мастер переговоров / SUCCESSFUL_SELLER** — at least **70%** of this year's
  published listings reached a completed sale (with at least one publication so
  the ratio has a denominator).
- **Маяк стабильности / CONSISTENT_PUBLISHER** — at least **10 published listings**
  and at least **50%** seller conversion.
- The displayed seller conversion is `min(salesCompleted, listingsPublished) / listingsPublished`;
  the numerator is capped because a sale completed this year can refer to a listing
  created earlier, while the recap percentage must stay within 0–100%.

Thematic achievements still use the weighted interest signal
`views + favorites * 4 + purchases * 12`, require one volume threshold
(30 views, 8 favorites, or 3 purchases), and at least **20%** dominance in the
user's total thematic signal. The catalogue now includes **Недвижимость** with the
**Решительный шаг / DECISIVE_STEP** achievement. All thematic achievements are
now explicitly shareable through the public achievement allow-list; the public
share DTO still exposes only the achievement title, not raw metrics or internal IDs.

## Frontend integration

Generated TypeScript schemas and the service descriptor are committed at:

```text
frontend/src/gen/recap/v1/recap_pb.ts
```

Install the client runtime:

```powershell
npm install @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-web
```

Create a typed browser client:

```typescript
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { RecapService } from "./gen/recap/v1/recap_pb";

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
- TypeScript schemas and service descriptors in `frontend/src/gen/recap/v1`

## Configuration

| Variable | Default |
| --- | --- |
| `API_ADDRESS` | empty; falls back to `HTTP_ADDR`, then Render `PORT`, then `:8080` |
| `STORAGE_BACKEND` | `clickhouse` (`memory` in the Render Docker image) |
| `CLICKHOUSE_DSN` | `clickhouse://recap:recap@clickhouse:9000/recap` |
| `SEED_DEMO_DATA` | `true` |
| `PROFILES_PATH` | `seeds/profiles.json` |
| `SCENARIOS_PATH` | `seeds/scenarios.json` |
| `STATIC_DIR` | `frontend/public` |
| `FRONTEND_DIR` | empty (set to `/app/web` in the Render Docker image) |
| `CORS_ALLOWED_ORIGINS` | localhost ports 3000 and 5173 |
| `AI_NARRATIVE_PROVIDER` | `auto` (`ollama` when local server + model are available, otherwise deterministic copy) |
| `OLLAMA_MODEL` | `qwen3:4b` |
| `OLLAMA_BASE_URL` | `http://localhost:11434` (`http://host.docker.internal:11434` in Docker Compose) |
| `OLLAMA_KEEP_ALIVE` | `5m` |
| `AI_NARRATIVE_TIMEOUT` | `20s` |
| `AI_NARRATIVE_MAX_CONCURRENCY` | `2` |
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
require a running ClickHouse instance. Start only the database before running them:

```powershell
docker compose up -d clickhouse
docker compose ps clickhouse
go test -tags=integration ./...
```

The default test DSN is `clickhouse://recap:recap@localhost:9000/recap`; override it
with `CLICKHOUSE_TEST_DSN` when ClickHouse is exposed elsewhere. A connection-refused
error on port `9000` means the integration dependency is not running, not that a unit
test failed.

Golden recap examples for frontend mocks are available under
`testdata/golden/`.

## Docker

```powershell
docker compose up --build
```