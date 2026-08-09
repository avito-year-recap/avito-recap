import { Navigate, Route, Routes } from "react-router-dom";
import { ActionDemoPage } from "../pages/action-demo/ActionDemoPage";
import { GenerationPage } from "../pages/generation/GenerationPage";
import { NotFoundPage } from "../pages/not-found/NotFoundPage";
import { ProfilesPage } from "../pages/profiles/ProfilesPage";
import { RecapPage } from "../pages/recap/RecapPage";
import { SharePage } from "../pages/share/SharePage";

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<ProfilesPage />} />
      <Route path="/generate/:profileCode" element={<GenerationPage />} />
      <Route path="/recap/:profileCode" element={<RecapPage />} />
      <Route path="/share/:shareId" element={<SharePage />} />
      <Route path="/demo/action/:actionCode" element={<ActionDemoPage />} />
      <Route path="/404" element={<NotFoundPage />} />
      <Route path="*" element={<Navigate to="/404" replace />} />
    </Routes>
  );
}
