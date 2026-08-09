import { describe, expect, it } from "vitest";
import { getRecapTotemStage } from "../shared/ui/year-totem-utils";

describe("getRecapTotemStage", () => {
  it("grows the visual code as the recap progresses", () => {
    expect(getRecapTotemStage(0, 9)).toBe(0);
    expect(getRecapTotemStage(2, 9)).toBe(1);
    expect(getRecapTotemStage(4, 9)).toBe(2);
    expect(getRecapTotemStage(6, 9)).toBe(3);
    expect(getRecapTotemStage(8, 9)).toBe(4);
  });

  it("uses the complete state for a single-card story", () => {
    expect(getRecapTotemStage(0, 1)).toBe(4);
  });
});
