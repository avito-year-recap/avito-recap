import type { PropsWithChildren } from "react";
import { BrowserRouter } from "react-router-dom";
import { QueryProvider } from "./QueryProvider";

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <BrowserRouter>
      <QueryProvider>{children}</QueryProvider>
    </BrowserRouter>
  );
}
