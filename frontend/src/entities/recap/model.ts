export type RecapCardType =
  | "INTRO"
  | "YEAR_ACTIVITY"
  | "TOP_CATEGORY"
  | "ACTIVE_MONTH"
  | "BEHAVIOR"
  | "ACHIEVEMENT"
  | "MISSED_OPPORTUNITY"
  | "NEXT_ACTION"
  | "SHARE";
export type ActionCode =
  | "FINISH_DRAFT"
  | "OPEN_FAVORITES"
  | "IMPROVE_LISTINGS"
  | "CONTINUE_DIALOGS"
  | "OPEN_TOP_CATEGORY"
  | "CREATE_LISTING"
  | "SAVE_SEARCH"
  | "VIEW_SIMILAR_LISTINGS"
  | "EXPLORE_RECOMMENDATIONS";
export type ActionTarget =
  | { type: "listing"; listingId: string }
  | { type: "dialog"; dialogId: string }
  | { type: "category"; categoryCode: string }
  | { type: "search"; categoryCode: string }
  | { type: "route"; path: string };

export interface Profile {
  name: string;
  description: string;
  avatarUrl: string;
  profileCode: string;
  tags: string[];
  accent: "blue" | "green" | "purple" | "coral";
}
interface BaseCard {
  id: string;
  position: number;
  title: string;
  description: string;
  explanation?: string;
  shareable: boolean;
}
export interface IntroCard extends BaseCard {
  type: "INTRO";
}
export interface YearActivityCard extends BaseCard {
  type: "YEAR_ACTIVITY";
  payload: {
    totalEvents: number;
    searches: number;
    totalViews: number;
    favoritesAdded: number;
    chatsStarted: number;
    listingsPublished: number;
    purchasesCompleted: number;
    salesCompleted: number;
  };
}
export interface TopCategoryCard extends BaseCard {
  type: "TOP_CATEGORY";
  payload: { categoryCode: string; category: string; categoryViews: number };
}
export interface ActiveMonthCard extends BaseCard {
  type: "ACTIVE_MONTH";
  payload: { month: number };
}
export interface BehaviorEvidence {
  metric: string;
  label: string;
  actualValue: number;
  threshold: number;
  comparison: "GTE" | "LTE";
  points: number;
  explanation: string;
}
export interface BehaviorCard extends BaseCard {
  type: "BEHAVIOR";
  payload: { code: string; score: number; evidence: BehaviorEvidence[] };
}
export interface AchievementCard extends BaseCard {
  type: "ACHIEVEMENT";
  payload: { codes: string[] };
}
export interface MissedOpportunityCard extends BaseCard {
  type: "MISSED_OPPORTUNITY";
  payload: {
    code: Extract<ActionCode, "SAVE_SEARCH" | "FINISH_DRAFT">;
    target: ActionTarget;
  };
}
export interface NextActionCard extends BaseCard {
  type: "NEXT_ACTION";
  payload: { code: ActionCode; target: ActionTarget };
}
export interface PublicSharePayload {
  shareId: string;
  year: number;
  behaviorTitle: string;
  achievementTitle?: string;
  topCategory?: string;
}
export interface ShareCard extends BaseCard {
  type: "SHARE";
  shareable: true;
  payload: PublicSharePayload;
}
export type RecapCard =
  | IntroCard
  | YearActivityCard
  | TopCategoryCard
  | ActiveMonthCard
  | BehaviorCard
  | AchievementCard
  | MissedOpportunityCard
  | NextActionCard
  | ShareCard;
export interface Achievement {
  code: string;
  title: string;
  reason: string;
  shareable: boolean;
}
export interface NextAction {
  code: ActionCode;
  title: string;
  description: string;
  explanation: string;
  buttonText: string;
  target: ActionTarget;
}
export interface Recap {
  id: string;
  year: number;
  ruleVersion: string;
  profile: Profile;
  cards: RecapCard[];
  achievements: Achievement[];
  nextAction: NextAction;
}
