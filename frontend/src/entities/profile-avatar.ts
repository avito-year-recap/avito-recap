import activeBuyerAvatar from "../assets/profile-avatars/active-buyer.png";
import activeSellerAvatar from "../assets/profile-avatars/active-seller.png";
import researcherAvatar from "../assets/profile-avatars/researcher.png";
import universalUserAvatar from "../assets/profile-avatars/universal-user.png";
import draftSellerAvatar from "../assets/profile-avatars/draft-seller.png";
import decisiveBuyerAvatar from "../assets/profile-avatars/decisive-buyer.png";
import sellerBuyerHybridAvatar from "../assets/profile-avatars/seller-buyer-hybrid.png";
import returningPublisherAvatar from "../assets/profile-avatars/returning-publisher.png";
import categoryBrowserAvatar from "../assets/profile-avatars/category-browser.png";
import recommendationNewcomerAvatar from "../assets/profile-avatars/recommendation-newcomer.png";
import steadyBuyerAvatar from "../assets/profile-avatars/steady-buyer.png";
import bookCollectorAvatar from "../assets/profile-avatars/book-collector.png";
import privateStyleHunterAvatar from "../assets/profile-avatars/private-style-hunter.png";
import makerWithDraftAvatar from "../assets/profile-avatars/maker-with-draft.png";
import petThresholdBuyerAvatar from "../assets/profile-avatars/pet-threshold-buyer.png";
import musicTravelerAvatar from "../assets/profile-avatars/music-traveler.png";
import listingRestartAvatar from "../assets/profile-avatars/listing-restart.png";

const bundledAvatarByProfileCode: Record<string, string> = {
  "active-buyer": activeBuyerAvatar,
  "active-seller": activeSellerAvatar,
  researcher: researcherAvatar,
  "universal-user": universalUserAvatar,
  "draft-seller": draftSellerAvatar,
  "decisive-buyer": decisiveBuyerAvatar,
  "seller-buyer-hybrid": sellerBuyerHybridAvatar,
  "returning-publisher": returningPublisherAvatar,
  "category-browser": categoryBrowserAvatar,
  "recommendation-newcomer": recommendationNewcomerAvatar,
  "steady-buyer": steadyBuyerAvatar,
  "book-collector": bookCollectorAvatar,
  "private-style-hunter": privateStyleHunterAvatar,
  "maker-with-draft": makerWithDraftAvatar,
  "pet-threshold-buyer": petThresholdBuyerAvatar,
  "music-traveler": musicTravelerAvatar,
  "listing-restart": listingRestartAvatar,
};

/**
 * Seed-profile avatars are bundled by Vite so production gets cache-busted,
 * deploy-safe asset URLs. The backend avatarUrl remains a fallback for
 * profiles that are not part of the built-in demo set.
 */
export function resolveProfileAvatarUrl(
  profileCode: string,
  backendAvatarUrl: string,
): string {
  return bundledAvatarByProfileCode[profileCode] ?? backendAvatarUrl;
}
