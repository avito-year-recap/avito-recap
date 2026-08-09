import { useMutation, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useEffect, useMemo, useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import { generateRecap } from "../../shared/api/recap-api";
import { ErrorState } from "../../shared/ui/AsyncState";
import { PageShell } from "../../shared/ui/PageShell";
import "./GenerationPage.css";

const steps = [
  "Собираем важные моменты года",
  "Складываем год в цифрах",
  "Находим главную тему",
  "Подбираем ачивки",
  "Готовим следующий шаг",
];

export function GenerationPage() {
  const { profileCode } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeStep, setActiveStep] = useState(0);
  const mutation = useMutation({
    mutationFn: generateRecap,
    onSuccess: (recap) =>
      queryClient.setQueryData(["recap", recap.profile.profileCode], recap),
  });

  useEffect(() => {
    if (!profileCode || mutation.isPending || mutation.isSuccess) return;
    mutation.mutate(profileCode);
  }, [mutation, profileCode]);

  useEffect(() => {
    if (mutation.isError) return;
    const timer = window.setInterval(
      () => setActiveStep((current) => Math.min(steps.length - 1, current + 1)),
      620,
    );
    return () => window.clearInterval(timer);
  }, [mutation.isError]);

  const isReady = mutation.isSuccess && activeStep === steps.length - 1;
  const recap = mutation.data;
  useEffect(() => {
    if (!isReady || !recap) return;
    const timer = window.setTimeout(
      () => navigate(`/recap/${recap.profile.profileCode}`, { replace: true }),
      520,
    );
    return () => window.clearTimeout(timer);
  }, [isReady, navigate, recap]);

  const progress = useMemo(
    () => Math.round(((activeStep + 1) / steps.length) * 100),
    [activeStep],
  );
  if (!profileCode) return <Navigate to="/" replace />;
  if (mutation.isError)
    return (
      <PageShell>
        <ErrorState
          title="Не удалось собрать итоги"
          description="Не удалось получить данные с сервера."
          onRetry={() => mutation.mutate(profileCode)}
        />
      </PageShell>
    );

  return (
    <PageShell compactHeader>
      <section className="generation-screen" aria-live="polite">
        <div className="generation-copy">
          <span className="generation-copy__eyebrow">Шаг 2 из 2</span>
          <h1>Собираем твою историю года…</h1>
          <p>
            Это займёт всего пару секунд: соберём самое важное и превратим активность в связную историю.
          </p>
          <div
            className="generation-progress"
            aria-label={`Готово на ${progress}%`}
          >
            <div
              className="generation-progress__value"
              style={{ width: `${progress}%` }}
            />
          </div>
          <ol className="generation-steps">
            {steps.map((step, index) => {
              const complete = index < activeStep || isReady;
              const active = index === activeStep && !isReady;
              return (
                <li
                  key={step}
                  className={`${complete ? "is-complete" : ""} ${active ? "is-active" : ""}`}
                >
                  <span aria-hidden="true">{complete ? "✓" : index + 1}</span>
                  <strong>{step}</strong>
                </li>
              );
            })}
          </ol>
        </div>
        <div className="generation-visual" aria-hidden="true">
          <motion.div
            className="generation-box"
            animate={{ y: [0, -8, 0] }}
            transition={{ repeat: Infinity, duration: 3.2, ease: "easeInOut" }}
          >
            <div className="generation-box__lid" />
            <div className="generation-card generation-card--one">♡</div>
            <div className="generation-card generation-card--two">⌕</div>
            <div className="generation-card generation-card--three">✓</div>
            <div className="generation-box__body">Avito</div>
          </motion.div>
          <span className="spark spark--one">✦</span>
          <span className="spark spark--two">✦</span>
          <span className="spark spark--three">•</span>
        </div>
      </section>
    </PageShell>
  );
}
