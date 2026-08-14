import { useMutation, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useEffect, useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import { generateRecap } from "../../shared/api/recap-api";
import { ErrorState } from "../../shared/ui/AsyncState";
import { PageShell } from "../../shared/ui/PageShell";
import "./GenerationPage.css";

const MIN_LOADING_TIME_MS = 2200;

export function GenerationPage() {
  const { profileCode } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [loadingReadyProfileCode, setLoadingReadyProfileCode] = useState<string | null>(null);
  const mutation = useMutation({
    mutationFn: generateRecap,
    onSuccess: (recap) =>
      queryClient.setQueryData(["recap", recap.profile.profileCode], recap),
  });

  useEffect(() => {
    if (!profileCode) return;

    const timer = window.setTimeout(
      () => setLoadingReadyProfileCode(profileCode),
      MIN_LOADING_TIME_MS,
    );

    return () => window.clearTimeout(timer);
  }, [profileCode]);

  const minimumLoadingTimePassed =
    Boolean(profileCode) && loadingReadyProfileCode === profileCode;

  useEffect(() => {
    if (!profileCode || mutation.isPending || mutation.isSuccess) return;
    mutation.mutate(profileCode);
  }, [mutation, profileCode]);

  const recap = mutation.data;

  useEffect(() => {
    if (!mutation.isSuccess || !recap || !minimumLoadingTimePassed) return;
    navigate(`/recap/${recap.profile.profileCode}`, { replace: true });
  }, [minimumLoadingTimePassed, mutation.isSuccess, navigate, recap]);

  if (!profileCode) return <Navigate to="/year" replace />;

  if (mutation.isError) {
    return (
      <PageShell fitViewport backTo="/account" backLabel="В кабинет">
        <ErrorState
          title="Не получилось открыть итоги"
          description="Попробуй ещё раз."
          onRetry={() => mutation.mutate(profileCode)}
        />
      </PageShell>
    );
  }

  return (
    <PageShell compactHeader fitViewport backTo="/account" backLabel="В кабинет">
      <section className="generation-screen" aria-live="polite">
        <div className="generation-copy">
          <span className="generation-copy__eyebrow">Твой 2025</span>
          <h1>Собираем твою историю года</h1>
          <p>Подбираем самые заметные моменты твоего 2025 на Авито.</p>

          <div className="generation-progress" aria-label="Собираем историю года">
            <span className="generation-progress__value" />
          </div>
        </div>

        <div className="generation-visual" aria-hidden="true">
          <motion.div
            className="generation-story-card"
            animate={{ y: [0, -6, 0], rotate: [-1.5, -0.5, -1.5] }}
            transition={{ repeat: Infinity, duration: 3.6, ease: "easeInOut" }}
          >
            <div className="generation-story-card__topline">
              <span>Мой год на Авито</span>
              <b>2025</b>
            </div>
            <strong className="generation-story-card__number">28</strong>
            <span className="generation-story-card__caption">сохранённых находок</span>
            <div className="generation-story-card__secret">
              <small>Твоя главная тема</small>
              <b>████████████</b>
            </div>
            <div className="generation-story-card__secret generation-story-card__secret--second">
              <small>Твой тип года</small>
              <b>██████████████</b>
            </div>
            <span className="generation-story-card__orb generation-story-card__orb--blue" />
            <span className="generation-story-card__orb generation-story-card__orb--green" />
            <span className="generation-story-card__orb generation-story-card__orb--purple" />
          </motion.div>
        </div>
      </section>
    </PageShell>
  );
}
