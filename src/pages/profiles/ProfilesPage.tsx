import { useQuery } from "@tanstack/react-query";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { resolveProfileAvatarUrl } from "../../entities/profile-avatar";
import type { Profile } from "../../entities/recap/model";
import { listProfiles } from "../../shared/api/recap-api";
import { getActiveProfile, setActiveProfile } from "../../shared/lib/active-profile";
import { ErrorState, PageLoader } from "../../shared/ui/AsyncState";
import { AppLogo } from "../../shared/ui/AppLogo";
import "./ProfilesPage.css";

const SAFE_RETURN_PATHS = new Set(["/", "/account"]);

const HERO_POSITIONS = ["one", "two", "three", "four"] as const;

function ProfileCard({
  profile,
  isActive,
  onSelect,
}: {
  profile: Profile;
  isActive: boolean;
  onSelect: () => void;
}) {
  const avatarUrl = resolveProfileAvatarUrl(profile.profileCode, profile.avatarUrl);

  return (
    <button
      type="button"
      className={`profile-choice profile-choice--${profile.accent}${isActive ? " is-active" : ""}`}
      onClick={onSelect}
      aria-current={isActive ? "true" : undefined}
      aria-label={isActive ? `${profile.name} — текущий профиль` : `Переключиться на профиль ${profile.name}`}
    >
      <span className="profile-choice__avatar-wrap" aria-hidden="true">
        <img className="profile-choice__avatar" src={avatarUrl} alt="" />
      </span>

      <span className="profile-choice__body">
        <span className="profile-choice__name-row">
          <strong>{profile.name}</strong>
          {isActive && <span className="profile-choice__current">Сейчас</span>}
        </span>
        <span className="profile-choice__description">{profile.description}</span>
      </span>

      <span className="profile-choice__action" aria-hidden="true">
        <i className={`hgi hgi-stroke ${isActive ? "hgi-user" : "hgi-arrow-right-01"}`} />
      </span>
    </button>
  );
}

export function ProfilesPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const profilesQuery = useQuery({ queryKey: ["profiles"], queryFn: listProfiles });
  const activeProfile = getActiveProfile();

  const requestedReturn = searchParams.get("return") ?? "/account";
  const returnTo = SAFE_RETURN_PATHS.has(requestedReturn) ? requestedReturn : "/account";

  const handleSelect = (profile: Profile) => {
    setActiveProfile(profile);
    navigate(returnTo, { replace: true });
  };

  const heroProfiles = profilesQuery.data?.slice(0, HERO_POSITIONS.length) ?? [];

  return (
    <div className="profiles-page">
      <main className="profiles-layout">
        <section className="profiles-hero" aria-labelledby="profiles-title">
          <Link to="/" className="profiles-hero__brand" aria-label="На главную Авито">
            <AppLogo />
          </Link>

          <div className="profiles-hero__copy">
            <span className="profiles-hero__eyebrow">Персональная история года</span>
            <h1 id="profiles-title">
              Твой год
              <br />
              <span>на Авито</span>
            </h1>
            <p>Выбери профиль и открой его персональные итоги года.</p>
          </div>

          <div className="profiles-hero__visual" aria-hidden="true">
            <div className="profiles-orbit profiles-orbit--outer" />
            <div className="profiles-orbit profiles-orbit--inner" />
            <div className="profiles-year">2025</div>

            {heroProfiles.map((profile, index) => (
              <span
                key={profile.profileCode}
                className={`profiles-floating profiles-floating--${HERO_POSITIONS[index]}`}
              >
                <img src={resolveProfileAvatarUrl(profile.profileCode, profile.avatarUrl)} alt="" />
              </span>
            ))}
          </div>
        </section>

        <section className="profiles-panel" aria-label="Выбор профиля">
          <div className="profiles-panel__heading">
            <div className="profiles-panel__heading-copy">
              <span>{profilesQuery.data ? `${profilesQuery.data.length} профилей` : "Профили"}</span>
              <h2>Выберите профиль</h2>
              <p>Он станет активным сразу после выбора.</p>
            </div>

            <Link to={returnTo} className="profiles-panel__back" aria-label="Вернуться назад">
              <i className="hgi hgi-stroke hgi-arrow-left-01" aria-hidden="true" />
            </Link>
          </div>

          {profilesQuery.isPending && (
            <div className="profiles-panel__state">
              <PageLoader label="Загружаем профили" />
            </div>
          )}

          {profilesQuery.isError && (
            <div className="profiles-panel__state">
              <ErrorState
                title="Не удалось загрузить профили"
                description="Попробуй ещё раз через пару секунд."
                onRetry={() => profilesQuery.refetch()}
              />
            </div>
          )}

          {profilesQuery.data && (
            <div className="profiles-list" role="list">
              {profilesQuery.data.map((profile) => (
                <ProfileCard
                  key={profile.profileCode}
                  profile={profile}
                  isActive={profile.profileCode === activeProfile.profileCode}
                  onSelect={() => handleSelect(profile)}
                />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
