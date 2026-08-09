import type { NextAction } from "../../entities/recap/model";

export function buildMockActionUrl(nextAction: NextAction): string {
  const params = new URLSearchParams();
  const target = nextAction.target;

  switch (target.type) {
    case "listing":
      params.set("listingId", target.listingId);
      break;
    case "dialog":
      params.set("dialogId", target.dialogId);
      break;
    case "category":
      params.set("categoryCode", target.categoryCode);
      break;
    case "search":
      params.set("categoryCode", target.categoryCode);
      break;
    case "route":
      params.set("path", target.path);
      break;
  }

  const query = params.toString();
  return `/demo/action/${nextAction.code}${query ? `?${query}` : ""}`;
}
