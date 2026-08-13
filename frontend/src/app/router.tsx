import { Navigate, Route, Routes } from "react-router-dom";
import { AccountPage } from "../pages/account/AccountPage";
import { ActionDemoPage } from "../pages/action-demo/ActionDemoPage";
import { AvitoHomePage } from "../pages/avito-home/AvitoHomePage";
import { GenerationPage } from "../pages/generation/GenerationPage";
import { NotFoundPage } from "../pages/not-found/NotFoundPage";
import { ProfilesPage } from "../pages/profiles/ProfilesPage";
import { RecapPage } from "../pages/recap/RecapPage";
import { SharePage } from "../pages/share/SharePage";
import { getActiveProfileCode } from "../shared/lib/active-profile";

function YearEntryRoute() {
  return <Navigate to={`/generate/${getActiveProfileCode()}`} replace />;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<AvitoHomePage />} />
      <Route path="/account" element={<AccountPage />} />
      <Route path="/year" element={<YearEntryRoute />} />
      <Route path="/profiles" element={<ProfilesPage />} />
      <Route path="/generate/:profileCode" element={<GenerationPage />} />
      <Route path="/recap/:profileCode" element={<RecapPage />} />
      <Route path="/share/:shareId" element={<SharePage />} />
      <Route path="/demo/action/:actionCode" element={<ActionDemoPage />} />
      <Route path="/404" element={<NotFoundPage />} />
      <Route path="*" element={<Navigate to="/404" replace />} />
    </Routes>
  );
}
