import { describe, expect, it } from "vitest";
import type { NextAction } from "../entities/recap/model";
import { buildMockActionUrl } from "../features/next-action/executeMockAction";
import { getActionVisual } from "../shared/lib/visual-registry";

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

  it("supports the first-listing action used by the starter seller recap", () => {
    const action: NextAction = {
      code: "CREATE_FIRST_LISTING",
      title: "Опубликуй первое объявление",
      description: "",
      explanation: "",
      buttonText: "Создать объявление",
      target: { type: "route", path: "/listings/new" },
    };

    expect(buildMockActionUrl(action)).toBe(
      "/demo/action/CREATE_FIRST_LISTING?path=%2Flistings%2Fnew",
    );
    expect(getActionVisual(action.code)).toMatchObject({
      tone: "coral",
      caption: "Создать первое объявление",
    });
  });

  it("uses a safe visual fallback for an unknown runtime action code", () => {
    expect(getActionVisual("FUTURE_ACTION_CODE")).toMatchObject({
      tone: "blue",
      caption: "Твой год на Авито",
    });
  });

});
