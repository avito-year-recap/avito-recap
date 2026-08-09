import type {
  ActionCode,
  ActionTarget,
  RecapCard,
} from "../../entities/recap/model";

export interface BackendRecapResponse {
  profile: {
    name: string;
    description: string;
    avatarUrl: string;
    profileCode: string;
  };
  recap: {
    id: string;
    year: number;
    ruleVersion: string;
    cards: RecapCard[];
    achievements: {
      code: string;
      title: string;
      reason: string;
      shareable: boolean;
    }[];
    nextAction: {
      code: ActionCode;
      title: string;
      description: string;
      explanation: string;
      buttonText: string;
      target: ActionTarget;
    };
  };
}
