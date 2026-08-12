# Generated Protobuf client

The source contract is `proto/recap/v1/recap.proto`. `buf generate` writes the
TypeScript client descriptors to `frontend/src/gen/recap/v1` and the Go
messages/Connect handlers to `gen/go/recap/v1`.

Run from the repository root:

```bash
buf generate
```

Do not edit generated `*_pb.ts` or `*.pb.go` files manually. The frontend uses
the generated `RecapService` descriptor through `createRecapTransport()` and
calls the real Go backend; mock data remains test/demo support only.
