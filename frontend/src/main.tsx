import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import { AppProviders } from "./app/providers/AppProviders";
import "./shared/styles/tokens.css";
import "./shared/styles/globals.css";

const root = document.getElementById("root");
if (!root) throw new Error("Root element was not found");

createRoot(root).render(
  <StrictMode>
    <AppProviders>
      <App />
    </AppProviders>
  </StrictMode>,
);
