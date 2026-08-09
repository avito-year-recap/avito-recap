import { describe, expect, it } from "vitest";
import { mapRecapResponse } from "../entities/recap/mapper";
import { mockRecaps } from "../shared/api/mock-data";

describe("mapRecapResponse", () => {
  it("sorts cards by position", () => {
    const response = structuredClone(mockRecaps[0]);
    response.recap.cards.reverse();
    expect(
      mapRecapResponse(response).cards.map((card) => card.position),
    ).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9]);
  });
  it("keeps only SHARE shareable", () => {
    const result = mapRecapResponse(structuredClone(mockRecaps[0]));
    expect(result.cards.filter((card) => card.shareable)).toHaveLength(1);
    expect(result.cards.at(-1)?.type).toBe("SHARE");
  });
  it("supports a five-card story", () => {
    const result = mapRecapResponse(structuredClone(mockRecaps[3]));
    expect(result.cards.map((card) => card.type)).toEqual([
      "INTRO",
      "YEAR_ACTIVITY",
      "BEHAVIOR",
      "NEXT_ACTION",
      "SHARE",
    ]);
  });
});

import { getPublicShare } from "../shared/api/recap-api";

describe("public share endpoint", () => {
  it("returns only the minimal public payload", async () => {
    const payload = await getPublicShare("share-marina-2025");

    expect(payload).toEqual({
      shareId: "share-marina-2025",
      year: 2025,
      behaviorTitle: "Глубокое исследование",
      achievementTitle: "Внимательное сравнение",
      topCategory: "Дом и дача",
    });
    expect("profile" in payload).toBe(false);
    expect("cards" in payload).toBe(false);
  });
});
