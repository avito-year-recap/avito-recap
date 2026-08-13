import {
  AchievementCode,
  ActionCode,
  BehaviorCode,
  EvidenceComparison,
} from "../../gen/recap/v1/recap_pb";
import type {
  Achievement as ProtoAchievement,
  ActionTarget as ProtoActionTarget,
  NextAction as ProtoNextAction,
  Profile as ProtoProfile,
  RecapCard as ProtoRecapCard,
  RecapResponse as ProtoRecapResponse,
} from "../../gen/recap/v1/recap_pb";
import { getProfileDisplayDescription, getProfilePresentation } from "../profile-presentation";
import { resolveProfileAvatarUrl } from "../profile-avatar";
import type {
  Achievement,
  ActionTarget,
  NextAction,
  Profile,
  PublicSharePayload,
  Recap,
  RecapCard,
} from "./model";

// protoc-gen-es emits enums as `as const` objects (erasable_syntax=true, see
// buf.gen.yaml) rather than real TS enums, so there is no built-in reverse
// (numeric value -> name) lookup like `SomeEnum[value]` would give.
function reverseLookup<T extends Record<string, number>>(
  source: T,
  value: number,
): keyof T {
  for (const key in source) {
    if (source[key] === value) return key;
  }
  throw new Error(`unknown enum value ${value}`);
}

interface ShareLikePayload {
  shareId: string;
  year: number;
  behaviorTitle: string;
  achievementTitle?: string;
  topCategory?: string;
}

export function profileFromProto(value: ProtoProfile): Profile {
  return {
    name: value.name,
    description: getProfileDisplayDescription(value.profileCode, value.description),
    avatarUrl: resolveProfileAvatarUrl(value.profileCode, value.avatarUrl),
    profileCode: value.profileCode,
    ...getProfilePresentation(value.profileCode),
  };
}

export function profilesFromProto(values: ProtoProfile[]): Profile[] {
  return values.map(profileFromProto);
}

function actionTargetFromProto(value: ProtoActionTarget | undefined): ActionTarget {
  switch (value?.target.case) {
    case "listing":
      return { type: "listing", listingId: value.target.value.listingId };
    case "dialog":
      return { type: "dialog", dialogId: value.target.value.dialogId };
    case "category":
      return { type: "category", categoryCode: value.target.value.categoryCode };
    case "search":
      return { type: "search", categoryCode: value.target.value.categoryCode };
    case "route":
      return { type: "route", path: value.target.value.path };
    default:
      throw new Error("action target is missing a payload case");
  }
}

function achievementFromProto(value: ProtoAchievement): Achievement {
  return {
    code: reverseLookup(AchievementCode, value.code) as string,
    title: value.title,
    reason: value.reason,
    shareable: value.shareable,
  };
}

function nextActionFromProto(value: ProtoNextAction): NextAction {
  return {
    code: reverseLookup(ActionCode, value.code) as NextAction["code"],
    title: value.title,
    description: value.description,
    explanation: value.explanation,
    buttonText: value.buttonText,
    target: actionTargetFromProto(value.target),
  };
}

export function publicShareFromProto(value: ShareLikePayload): PublicSharePayload {
  return {
    shareId: value.shareId,
    year: value.year,
    behaviorTitle: value.behaviorTitle,
    achievementTitle: value.achievementTitle,
    topCategory: value.topCategory,
  };
}

function cardFromProto(value: ProtoRecapCard): RecapCard {
  const base = {
    id: value.id,
    position: value.position,
    title: value.title,
    description: value.description,
    explanation: value.explanation || undefined,
    shareable: value.shareable,
  };
  switch (value.payload.case) {
    case "intro":
      return { ...base, type: "INTRO" };
    case "yearActivity": {
      const payload = value.payload.value;
      return {
        ...base,
        type: "YEAR_ACTIVITY",
        payload: {
          totalEvents: Number(payload.totalEvents),
          searches: Number(payload.searches),
          totalViews: Number(payload.totalViews),
          favoritesAdded: Number(payload.favoritesAdded),
          chatsStarted: Number(payload.chatsStarted),
          listingsPublished: Number(payload.listingsPublished),
          purchasesCompleted: Number(payload.purchasesCompleted),
          salesCompleted: Number(payload.salesCompleted),
        },
      };
    }
    case "topCategory": {
      const payload = value.payload.value;
      return {
        ...base,
        type: "TOP_CATEGORY",
        payload: {
          categoryCode: payload.categoryCode,
          category: payload.category,
          categoryViews: Number(payload.categoryViews),
        },
      };
    }
    case "activeMonth":
      return {
        ...base,
        type: "ACTIVE_MONTH",
        payload: { month: value.payload.value.month },
      };
    case "behavior": {
      const payload = value.payload.value;
      return {
        ...base,
        type: "BEHAVIOR",
        payload: {
          code: reverseLookup(BehaviorCode, payload.code) as string,
          score: payload.score,
          evidence: payload.evidence.map((item) => ({
            metric: item.metric,
            label: item.label,
            actualValue: item.actualValue,
            threshold: item.threshold,
            comparison: reverseLookup(EvidenceComparison, item.comparison) as
              | "GTE"
              | "LTE",
            points: item.points,
            explanation: item.explanation,
          })),
        },
      };
    }
    case "achievement":
      return {
        ...base,
        type: "ACHIEVEMENT",
        payload: {
          codes: value.payload.value.codes.map(
            (code) => reverseLookup(AchievementCode, code) as string,
          ),
        },
      };
    case "missedOpportunity": {
      const payload = value.payload.value;
      return {
        ...base,
        type: "MISSED_OPPORTUNITY",
        payload: {
          code: reverseLookup(ActionCode, payload.code) as
            | "SAVE_SEARCH"
            | "FINISH_DRAFT",
          target: actionTargetFromProto(payload.target),
        },
      };
    }
    case "nextAction": {
      const payload = value.payload.value;
      return {
        ...base,
        type: "NEXT_ACTION",
        payload: {
          code: reverseLookup(ActionCode, payload.code) as NextAction["code"],
          target: actionTargetFromProto(payload.target),
        },
      };
    }
    case "share":
      return {
        ...base,
        type: "SHARE",
        shareable: true,
        payload: publicShareFromProto(value.payload.value),
      };
    default:
      throw new Error(`recap card ${value.id} is missing a payload`);
  }
}

export function recapFromProto(response: ProtoRecapResponse): Recap {
  if (!response.profile || !response.recap) {
    throw new Error("recap response is missing profile or recap");
  }
  if (!response.recap.nextAction) {
    throw new Error("recap response is missing next action");
  }
  return {
    id: response.recap.id,
    year: response.recap.year,
    ruleVersion: response.recap.ruleVersion,
    profile: profileFromProto(response.profile),
    cards: response.recap.cards
      .map(cardFromProto)
      .sort((a, b) => a.position - b.position),
    achievements: response.recap.achievements.map(achievementFromProto),
    nextAction: nextActionFromProto(response.recap.nextAction),
  };
}
