import { describe, expect, it } from "vitest";
import type { NextAction } from "../entities/recap/model";
import { buildMockActionUrl } from "../features/next-action/executeMockAction";

describe("buildMockActionUrl", () => {
  it("keeps a typed search target in the demo URL", () => {
    const action: NextAction = {
      code: "SAVE_SEARCH",
      title: "Сохранить поиск",
      description: "",
      explanation: "",
      buttonText: "Сохранить поиск",
      target: { type: "search", categoryCode: "HOME_AND_GARDEN" },
    };

    expect(buildMockActionUrl(action)).toBe(
      "/demo/action/SAVE_SEARCH?categoryCode=HOME_AND_GARDEN",
    );
  });

  it("keeps listing ids for listing-based actions", () => {
    const action: NextAction = {
      code: "FINISH_DRAFT",
      title: "Открыть черновик",
      description: "",
      explanation: "",
      buttonText: "Открыть черновик",
      target: { type: "listing", listingId: "draft-42" },
    };

    expect(buildMockActionUrl(action)).toBe(
      "/demo/action/FINISH_DRAFT?listingId=draft-42",
    );
  });
});
