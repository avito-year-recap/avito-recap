import { createClient } from "@connectrpc/connect";
import { RecapService } from "../../gen/recap/v1/recap_pb";
import {
  profilesFromProto,
  publicShareFromProto,
  recapFromProto,
} from "../../entities/recap/proto-mapper";
import type { Profile, PublicSharePayload, Recap } from "../../entities/recap/model";
import { createRecapTransport } from "./connect-transport";

// In the Render single-service build, Go exposes the Connect API under /api
// and serves the React bundle from the same origin. The Vite dev server still
// talks to the backend directly on localhost (CORS-allowed by default).
const API_BASE_URL = import.meta.env.DEV ? "http://localhost:8080" : "/api";

// Seed scenarios (seeds/scenarios.json) only cover 2025, the last calendar
// year completed before this demo's "now" — GenerateRecap/GetRecap reject
// any year that isn't fully in the past.
const RECAP_YEAR = 2025;

const client = createClient(RecapService, createRecapTransport(API_BASE_URL));

export async function listProfiles(): Promise<Profile[]> {
  const response = await client.listProfiles({});
  return profilesFromProto(response.profiles);
}

export async function generateRecap(profileCode: string): Promise<Recap> {
  const response = await client.generateRecap({
    profileCode,
    year: RECAP_YEAR,
  });
  return recapFromProto(response);
}

export async function getRecap(profileCode: string): Promise<Recap> {
  const response = await client.getRecap({
    profileCode,
    year: RECAP_YEAR,
  });
  return recapFromProto(response);
}

export async function getPublicShare(
  shareId: string,
): Promise<PublicSharePayload> {
  const response = await client.getPublicShare({ shareId });
  if (!response.share) throw new Error("SHARE_NOT_FOUND");
  return publicShareFromProto(response.share);
}
