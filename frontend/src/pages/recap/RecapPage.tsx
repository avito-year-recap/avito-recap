import { useQuery } from "@tanstack/react-query";
import { Navigate, useParams } from "react-router-dom";
import { getRecap } from "../../shared/api/recap-api";
import { ErrorState, PageLoader } from "../../shared/ui/AsyncState";
import { PageShell } from "../../shared/ui/PageShell";
import { RecapPlayer } from "../../widgets/recap-player/RecapPlayer";

export function RecapPage() {
  const { profileCode } = useParams();
  const query = useQuery({
    queryKey: ["recap", profileCode],
    queryFn: () => getRecap(profileCode ?? ""),
    enabled: Boolean(profileCode),
  });

  if (!profileCode) return <Navigate to="/" replace />;
  if (query.isPending)
    return (
      <PageShell compactHeader fitViewport backToProfiles>
        <PageLoader label="Открываем твою историю года" />
      </PageShell>
    );
  if (query.isError)
    return (
      <PageShell compactHeader fitViewport backToProfiles>
        <ErrorState
          title="Итоги не найдены"
          description="Для этого тестового профиля нет готового recap."
          onRetry={() => query.refetch()}
        />
      </PageShell>
    );

  return (
    <PageShell
      compactHeader
      fitViewport
      backToProfiles
      actions={<span className="demo-chip">{query.data.profile.name} · {query.data.year}</span>}
    >
      <RecapPlayer recap={query.data} />
    </PageShell>
  );
}
