import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Profile } from "../../entities/recap/model";
import { getRecap, listProfiles } from "../../shared/api/recap-api";
import { ErrorState, PageLoader } from "../../shared/ui/AsyncState";
import { Button } from "../../shared/ui/Button";
import { getProfileTeaser, readCompletedProfiles } from "../../shared/lib/experience-utils";
import { YearTotem } from "../../shared/ui/YearTotem";
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
  const teaser = getProfileTeaser(profile);
  return (
    <button
      type="button"
      className={`profile-card profile-card--${profile.accent} ${selected ? "profile-card--selected" : ""}`}
      onClick={onSelect}
      aria-pressed={selected}
      data-accent={profile.accent}
    >
      <span className="profile-card__signature" aria-hidden="true"><i /><i /><i /></span>
      <img src={profile.avatarUrl} alt="" className="profile-card__avatar" />
      <span className="profile-card__body">
        <span className="profile-card__topline">
          <strong>{profile.name}</strong>
          <span className="profile-card__status" aria-hidden="true">
            {selected ? "✓" : "→"}
          </span>
        </span>
        <span className="profile-card__description">{profile.description}</span>
        <span className="profile-card__tags">
          {profile.tags.map((tag) => (
            <span key={tag}>{tag}</span>
          ))}
        </span>
        <span className="profile-card__teaser">
          <i aria-hidden="true">✦</i>
          {teaser}
        </span>
      </span>
    </button>
  );
}

function CompletedComparison({ recaps }: { recaps: Awaited<ReturnType<typeof getRecap>>[] }) {
  if (recaps.length < 2) return null;
  const pair = recaps.slice(0, 2);
  return (
    <section className="profiles-comparison" aria-label="Сравнение просмотренных историй">
      <div className="profiles-comparison__heading"><span>После просмотра</span><h2>Насколько по-разному складывается год</h2><p>Мы не сравниваем людей по цифрам. Здесь видно только, как разные данные меняют визуальный язык recap.</p></div>
      <div className="profiles-comparison__pair">
        {pair.map((recap) => {
          const behavior = recap.cards.find((card) => card.type === "BEHAVIOR");
          return <article key={recap.id}><div><YearTotem recap={recap} stage={4} /></div><span>{recap.profile.name}</span><strong>{behavior?.title ?? "Разные сценарии"}</strong></article>;
        })}
        <i aria-hidden="true">≠</i>
      </div>
    </section>
  );
}

export function ProfilesPage() {
  const navigate = useNavigate();
  const profilesQuery = useQuery({
    queryKey: ["profiles"],
    queryFn: listProfiles,
  });
  const completedCodes = useMemo(() => readCompletedProfiles(), []);
  const comparisonQuery = useQuery({
    queryKey: ["completed-recaps", completedCodes],
    queryFn: () => Promise.all(completedCodes.slice(0, 2).map((code) => getRecap(code))),
    enabled: completedCodes.length >= 2,
  });
  const [selectedCode, setSelectedCode] = useState<string | null>(null);
  const selectedProfile = useMemo(
    () =>
      profilesQuery.data?.find(
        (profile) => profile.profileCode === selectedCode,
      ) ?? null,
    [profilesQuery.data, selectedCode],
  );

  if (profilesQuery.isPending)
    return (
      <PageShell>
        <PageLoader label="Загружаем тестовые профили" />
      </PageShell>
    );
  if (profilesQuery.isError)
    return (
      <PageShell>
        <ErrorState
          title="Не удалось загрузить профили"
          description="Не удалось получить тестовые профили. Попробуй ещё раз."
          onRetry={() => profilesQuery.refetch()}
        />
      </PageShell>
    );

  return (
    <PageShell
      actions={<span className="demo-chip">Демо · 5 разных историй</span>}
    >
      <section className="profiles-layout">
        <div className="profiles-hero">
          <span className="profiles-hero__eyebrow">
            Персональная история года
          </span>
          <h1>
            Твой год
            <br />
            <span>на Авито</span>
          </h1>
          <p>
            Твои поиски, просмотры, избранное и сделки складываются в короткую персональную историю — с выводами, ачивками и следующим шагом.
          </p>
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
            <span>✓ Только агрегированные данные</span>
            <span>✓ Без личных переписок</span>
            <span>✓ Делится только безопасный финал</span>
          </div>
        </div>
        <div className="profiles-panel">
          <div className="profiles-panel__heading">
            <span>Шаг 1 из 2</span>
            <h2>Выбери тестовый профиль</h2>
            <p>
              У каждого профиля свой ритм года. Наведи на карточку: интерфейс даст подсказку, но не раскроет финальный сценарий заранее.
            </p>
          </div>
          <div className="profiles-list">
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
              onClick={() =>
                selectedProfile &&
                navigate(`/generate/${selectedProfile.profileCode}`)
              }
            >
              {selectedProfile
                ? `Собрать итоги ${selectedProfile.name}`
                : "Сначала выбери профиль"}
            </Button>
            <small>Можно вернуться и посмотреть, как меняется история у другого профиля.</small>
          </div>
        </div>
      </section>
      {comparisonQuery.data && <CompletedComparison recaps={comparisonQuery.data} />}
    </PageShell>
  );
}
