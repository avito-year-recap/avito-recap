import { createConnectTransport } from "@connectrpc/connect-web";

/**
 * Готовая точка подключения сгенерированного ConnectRPC-клиента.
 * Пока UI использует recap-api.ts с моками. Protobuf-контракт лежит в
 * proto/recap/v1/recap.proto и уже сгенерирован в gen/ts/recap/v1;
 * этот transport передаётся в createClient(RecapService, transport).
 */
export function createRecapTransport(baseUrl: string) {
  return createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
}
