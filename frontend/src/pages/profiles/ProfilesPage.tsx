import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Profile } from "../../entities/recap/model";
import { listProfiles } from "../../shared/api/recap-api";
import { ErrorState, PageLoader } from "../../shared/ui/AsyncState";
import { Button } from "../../shared/ui/Button";
import { PageShell } from "../../shared/ui/PageShell";
import "./ProfilesPage.css";

function ProfileCard({
  profile,
  selected,
  onSelect,
}: {
  profile: Profile;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      className={`profile-card profile-card--${profile.accent} ${selected ? "profile-card--selected" : ""}`}
      onClick={onSelect}
      aria-pressed={selected}
    >
      <img src={profile.avatarUrl} alt="" className="profile-card__avatar" />
      <span className="profile-card__body">
        <span className="profile-card__topline">
          <strong>{profile.name}</strong>
          <span className="profile-card__status" aria-hidden="true">{selected ? "✓" : "→"}</span>
        </span>
        <span className="profile-card__description">{profile.description}</span>
      </span>
    </button>
  );
}

export function ProfilesPage() {
  const navigate = useNavigate();
  const profilesQuery = useQuery({ queryKey: ["profiles"], queryFn: listProfiles });
  const [selectedCode, setSelectedCode] = useState<string | null>(null);
  const selectedProfile = useMemo(
    () => profilesQuery.data?.find((profile) => profile.profileCode === selectedCode) ?? null,
    [profilesQuery.data, selectedCode],
  );

  if (profilesQuery.isPending) {
    return <PageShell fitViewport><PageLoader label="Загружаем тестовые профили" /></PageShell>;
  }
  if (profilesQuery.isError) {
    return (
      <PageShell fitViewport>
        <ErrorState
          title="Не удалось загрузить профили"
          description="Не удалось получить тестовые профили. Попробуй ещё раз."
          onRetry={() => profilesQuery.refetch()}
        />
      </PageShell>
    );
  }

  return (
    <PageShell
      fitViewport
      actions={<span className="demo-chip">{profilesQuery.data.length} тестовых профилей</span>}
    >
      <section className="profiles-layout">
        <div className="profiles-hero">
          <span className="profiles-hero__eyebrow">Персональная история года</span>
          <h1>Твой год<br /><span>на Авито</span></h1>
          <p>Выбери профиль и посмотри, как его активность складывается в персональные итоги года.</p>
          <div className="profiles-hero__visual" aria-hidden="true">
            <div className="orbit orbit--one" />
            <div className="orbit orbit--two" />
            <div className="year-orb">2025</div>
            <span className="floating-item floating-item--chair">🪑</span>
            <span className="floating-item floating-item--car">🚙</span>
            <span className="floating-item floating-item--phone">📱</span>
            <span className="floating-item floating-item--bike">🚲</span>
          </div>
          <div className="privacy-strip">
            <span>✓ Агрегированные данные</span>
            <span>✓ Без личных переписок</span>
            <span>✓ Публичен только безопасный финал</span>
          </div>
        </div>

        <div className="profiles-panel">
          <div className="profiles-panel__heading">
            <span>Шаг 1 из 2</span>
            <h2>Выбери тестовый профиль</h2>
            <p>Выбери профиль, чтобы собрать его персональные итоги.</p>
          </div>
          <div className="profiles-list" role="list">
            {profilesQuery.data.map((profile) => (
              <ProfileCard
                key={profile.profileCode}
                profile={profile}
                selected={profile.profileCode === selectedCode}
                onSelect={() => setSelectedCode(profile.profileCode)}
              />
            ))}
          </div>
          <div className="profiles-panel__footer">
            <Button
              fullWidth
              disabled={!selectedProfile}
              onClick={() => selectedProfile && navigate(`/generate/${selectedProfile.profileCode}`)}
            >
              {selectedProfile ? `Собрать итоги ${selectedProfile.name}` : "Сначала выбери профиль"}
            </Button>
          </div>
        </div>
      </section>
    </PageShell>
  );
}
