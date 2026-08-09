# Generated Protobuf client

This directory is generated from `proto/avito/recap/v1/recap.proto`.

Run:

```bash
npm run proto:generate
```

Do not edit generated `*_pb.ts` files manually.

The generated `RecapService` descriptor can be used with the existing
`createRecapTransport()` and ConnectRPC `createClient()` when the real backend
is connected. The current MVP keeps using mock data, so code generation is not
required to run the mock UI.
