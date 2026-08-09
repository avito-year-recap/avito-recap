import type { BackendRecapResponse } from "../../shared/api/backend-contract";
import { getProfilePresentation } from "../profile-presentation";
import type { Recap } from "./model";

export function mapRecapResponse(response: BackendRecapResponse): Recap {
  return {
    id: response.recap.id,
    year: response.recap.year,
    ruleVersion: response.recap.ruleVersion,
    profile: {
      ...response.profile,
      ...getProfilePresentation(response.profile.profileCode),
    },
    cards: structuredClone(response.recap.cards).sort(
      (a, b) => a.position - b.position,
    ),
    achievements: response.recap.achievements.map((item) => ({ ...item })),
    nextAction: structuredClone(response.recap.nextAction),
  };
}
