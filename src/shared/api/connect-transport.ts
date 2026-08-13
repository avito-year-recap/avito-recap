import { createConnectTransport } from "@connectrpc/connect-web";

/**
 * Protobuf-контракт лежит в proto/recap/v1/recap.proto, сгенерирован в
 * frontend/src/gen/recap/v1; этот transport передаётся в
 * createClient(RecapService, transport) в recap-api.ts.
 */
export function createRecapTransport(baseUrl: string) {
  return createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
}
