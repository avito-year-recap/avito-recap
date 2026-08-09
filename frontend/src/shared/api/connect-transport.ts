import { createConnectTransport } from "@connectrpc/connect-web";

/**
 * Готовая точка подключения сгенерированного ConnectRPC-клиента.
 * Пока UI использует recap-api.ts с моками. Protobuf-контракт уже лежит в
 * proto/avito/recap/v1/recap.proto; после `npm run proto:generate` этот
 * transport передаётся в createClient(RecapService, transport).
 */
export function createRecapTransport(baseUrl: string) {
  return createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
}
