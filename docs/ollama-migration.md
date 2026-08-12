# AI Narrative: migration from OpenAI to local Ollama

## Goal

Remove the paid cloud-API dependency from the narrative layer while keeping the
same domain boundary: the LLM may rewrite only existing card descriptions and
must not decide behavior, achievements, metrics or the next action.

## Architecture before

```text
application.Service
  -> narrative.Generator
  -> narrative.Limited
  -> narrative.BestEffort
  -> internal/ai/openai
  -> OpenAI /v1/responses
```

Configuration used `OPENAI_API_KEY`, `OPENAI_MODEL` and `OPENAI_BASE_URL`.

## Architecture now

```text
application.Service
  -> narrative.Enricher (BestEffort)
     -> narrative.Limited
        -> narrative.Generator (Ollama adapter)
        -> local Ollama /api/chat
```

The provider-facing `narrative.Generator` interface remains provider-independent.
The application now depends on the narrower `narrative.Enricher` boundary, so provider
errors and post-generation `Apply` validation errors are handled in one BestEffort
component instead of being split across layers.

## Changed files

- `internal/ai/openai/*` removed.
- `internal/ai/ollama/narrative.go` added: local Ollama HTTP adapter.
- `internal/ai/ollama/narrative_test.go` added: request/response and model-check tests.
- `internal/config/config.go` changed from OpenAI key/model/base URL to Ollama model/base URL/keep-alive.
- `internal/config/config_test.go` updated for Ollama settings.
- `cmd/api/main.go` now builds and probes the Ollama generator.
- `docker-compose.yml` points the API container to host Ollama through `host.docker.internal`.
- `render.yaml` explicitly disables local-only AI on the hosted demo.
- `README.md` documents local Ollama installation and startup.
- `internal/recap/narrative/limited.go` documentation now describes local inference resource protection.

## Ollama request

The adapter calls:

```text
POST {OLLAMA_BASE_URL}/api/chat
```

with:

- `stream=false` so the backend receives one response;
- `think=false` because hidden model reasoning is not needed for copywriting;
- a JSON Schema in `format` so the result is constrained to the exact number of AI-editable `cards[]` with `id` and `description`;
- `keep_alive` to avoid reloading the model for every recap;
- only privacy-safe aggregate recap facts and the explicit editable-card ID allow-list.

AI may rewrite `INTRO`, `YEAR_ACTIVITY`, `TOP_CATEGORY`, `ACTIVE_MONTH`, `BEHAVIOR`,
`ACHIEVEMENT`, `MISSED_OPPORTUNITY` and `NEXT_ACTION`. `SHARE` is never sent as an
editable card and remains deterministic. A non-empty AI result must cover every
editable card exactly once; partial/duplicate/unknown output is rejected atomically.

Before enabling the provider, startup checks the configured model with:

```text
POST {OLLAMA_BASE_URL}/api/show
```

`AI_NARRATIVE_PROVIDER=auto` disables AI when Ollama/model is unavailable.
`AI_NARRATIVE_PROVIDER=ollama` treats that condition as a startup configuration
error. Runtime inference failures are still handled by `narrative.BestEffort`,
so deterministic recap copy remains available.

## Configuration

```text
AI_NARRATIVE_PROVIDER=auto|ollama|off
OLLAMA_MODEL=qwen3:4b
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_KEEP_ALIVE=5m
AI_NARRATIVE_TIMEOUT=20s
AI_NARRATIVE_MAX_CONCURRENCY=2
```

There is no API key in the local Ollama path.

## Local startup

```powershell
ollama pull qwen3:4b
$env:AI_NARRATIVE_PROVIDER="ollama"
go run ./cmd/api
```

For Docker Compose, keep Ollama running on the host; the compose file uses
`http://host.docker.internal:11434` from inside the API container.
