# Protobuf / ConnectRPC contract

The frontend now contains an explicit Protobuf contract at:

`proto/avito/recap/v1/recap.proto`

It mirrors the recap data used by the UI:

- profile: `name`, `description`, `avatarUrl`, `profileCode`;
- ordered recap cards from `INTRO` through `SHARE`;
- the 8 `YEAR_ACTIVITY` counters;
- `TOP_CATEGORY` (`categoryCode`, `category`, `categoryViews`);
- `ACTIVE_MONTH` (`month` only — no invented monthly event count);
- `BEHAVIOR` with `code`, `score`, and `evidence`;
- `ACHIEVEMENT` codes plus full top-level achievement objects;
- `MISSED_OPPORTUNITY` and `NEXT_ACTION` with typed targets;
- `nextAction.buttonText` at recap level;
- a minimal public SHARE payload that contains no private metrics or object IDs.

## Why `oneof`

`RecapCard.payload` and `ActionTarget.target` are represented as Protobuf
`oneof`s. This gives generated TypeScript a discriminated shape and keeps the
frontend from accidentally reading, for example, a category payload as an
activity payload.

## Commands

```bash
npm run proto:lint
npm run proto:format
npm run proto:generate
```

Generation uses `@bufbuild/buf` and `@bufbuild/protoc-gen-es` already present in
the project and writes TypeScript into `src/shared/api/generated`.

The mock API remains the default data source. When a real ConnectRPC endpoint
is available, the generated `RecapService` descriptor should be passed to
`createClient()` together with `createRecapTransport()`.

## Frontend-only profile presentation

The profile chips and accent colors used by the demo are intentionally **not**
part of the backend contract. They are mapped locally by `profileCode` in
`src/entities/profile-presentation.ts`. This keeps the Proto aligned with the
provided backend specification while preserving the richer test-profile UI.
