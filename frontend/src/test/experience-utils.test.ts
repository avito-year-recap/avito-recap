import { describe, expect, it } from "vitest";
import type { BehaviorCard, YearActivityCard } from "../entities/recap/model";
import { deriveBehaviorTraits, getActionBeforeAfter, getDominantActivity, getSecretVisualBonus } from "../shared/lib/experience-utils";

describe("experience utils", () => {
  it("derives behavior portrait only from evidence", () => {
    const card: BehaviorCard = {
      id: "behavior",
      type: "BEHAVIOR",
      position: 1,
      title: "Глубокое исследование",
      description: "",
      shareable: false,
      payload: {
        code: "RESEARCHER",
        score: 80,
        evidence: [{ metric: "totalViews", label: "Просмотры", actualValue: 1284, threshold: 100, comparison: "GTE", points: 35, explanation: "Много просмотров" }],
      },
    };
    expect(deriveBehaviorTraits(card)[0]).toMatchObject({ title: "Сравниваешь", value: "1 284 · порог ≥ 100" });
  });

  it("finds dominant visible activity without external comparison", () => {
    const card: YearActivityCard = {
      id: "activity",
      type: "YEAR_ACTIVITY",
      position: 1,
      title: "",
      description: "",
      shareable: false,
      payload: { totalEvents: 1785, searches: 356, totalViews: 1284, favoritesAdded: 96, chatsStarted: 18, listingsPublished: 24, purchasesCompleted: 4, salesCompleted: 3 },
    };
    expect(getDominantActivity(card)).toMatchObject({ key: "totalViews", value: 1284, label: "Просмотры" });
  });

  it("has copy for the first-listing action", () => {
    expect(getActionBeforeAfter("CREATE_FIRST_LISTING")).toEqual({
      before: "Первое объявление ещё не опубликовано",
      after: "Первое объявление уже создаётся",
    });
  });

  it("marks secret result as a visual bonus, not backend achievement", () => {
    expect(getSecretVisualBonus(["MASTER_OF_FAVORITES", "BROAD_INTERESTS", "ATTENTIVE_RESEARCHER"])?.title).toBe("Куратор находок");
  });
});
