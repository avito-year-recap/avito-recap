import { getProfilePresentation } from "../../entities/profile-presentation";
import { mapRecapResponse } from "../../entities/recap/mapper";
import type {
  Profile,
  PublicSharePayload,
  Recap,
  ShareCard,
} from "../../entities/recap/model";
import { mockRecaps } from "./mock-data";

function wait(ms = 420) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, ms));
}
function clone<T>(value: T): T {
  return structuredClone(value);
}

export async function listProfiles(): Promise<Profile[]> {
  await wait(220);
  return mockRecaps.map((item) => ({
    ...clone(item.profile),
    ...getProfilePresentation(item.profile.profileCode),
  }));
}

export async function generateRecap(profileCode: string): Promise<Recap> {
  await wait(650);
  const response = mockRecaps.find(
    (item) => item.profile.profileCode === profileCode,
  );
  if (!response) throw new Error("PROFILE_NOT_FOUND");
  return mapRecapResponse(clone(response));
}

export async function getRecap(profileCode: string): Promise<Recap> {
  await wait();
  const response = mockRecaps.find(
    (item) => item.profile.profileCode === profileCode,
  );
  if (!response) throw new Error("RECAP_NOT_FOUND");
  return mapRecapResponse(clone(response));
}

export async function getPublicShare(
  shareId: string,
): Promise<PublicSharePayload> {
  await wait(300);
  const shareCard = mockRecaps
    .flatMap((item) => item.recap.cards)
    .find(
      (card): card is ShareCard =>
        card.type === "SHARE" && card.payload.shareId === shareId,
    );

  if (!shareCard) throw new Error("SHARE_NOT_FOUND");

  // Имитируем отдельный публичный RPC: наружу выходит только минимальный SHARE payload.
  return clone(shareCard.payload);
}
