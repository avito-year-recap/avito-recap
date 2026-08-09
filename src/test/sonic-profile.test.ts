import { describe, expect, it } from "vitest";
import { normalizeSonicProfile, sonicProfiles } from "../shared/sound/soundProfiles";

describe("sonic profiles", () => {
  it("keeps distinct sound characters for product behaviors", () => {
    expect(sonicProfiles.RESEARCHER.movement).toBe("glass");
    expect(sonicProfiles.ACTIVE_SELLER.movement).toBe("precise");
    expect(sonicProfiles.FIND_HUNTER.root).not.toBe(sonicProfiles.RESEARCHER.root);
  });

  it("falls back to universal profile for unknown future behavior", () => {
    expect(normalizeSonicProfile("FUTURE_BEHAVIOR")).toBe("UNIVERSAL_USER");
  });
});
