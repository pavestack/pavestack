import React, { useState } from "react";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { CatalogProvider } from "./CatalogContext";
import { ThemeProvider } from "./ThemeContext";
import { Sidebar } from "./layout/Sidebar";
import { Header } from "./layout/Header";
import { Overview } from "./routes/Overview";
import { ServiceDetail } from "./routes/ServiceDetail";
import { CreateServiceWizard } from "./routes/CreateService/CreateServiceWizard";
import { RequestAccess } from "./routes/RequestAccess";
import { Scorecards } from "./routes/Scorecards";
import { Observability } from "./routes/Observability";
import { Docs } from "./routes/Docs";
import { NotFound } from "./routes/NotFound";

function Shell() {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="app-shell">
      {sidebarOpen && (
        <div className="app-sidebar-backdrop" onClick={() => setSidebarOpen(false)} aria-hidden="true" />
      )}
      <Sidebar open={sidebarOpen} onNavigate={() => setSidebarOpen(false)} />
      <div className="app-main-col">
        <Header onToggleSidebar={() => setSidebarOpen((v) => !v)} />
        <main id="main-content" className="app-scroll-region">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 py-6">
            <Routes>
              <Route path="/" element={<Overview />} />
              <Route path="/services/:name" element={<ServiceDetail />} />
              <Route path="/create" element={<CreateServiceWizard />} />
              <Route path="/access" element={<RequestAccess />} />
              <Route path="/scorecards" element={<Scorecards />} />
              <Route path="/observability" element={<Observability />} />
              <Route path="/docs" element={<Docs />} />
              <Route path="/docs/:section" element={<Docs />} />
              <Route path="*" element={<NotFound />} />
            </Routes>
          </div>
        </main>
      </div>
    </div>
  );
}

export function App() {
  return (
    <ThemeProvider>
      <CatalogProvider>
        <BrowserRouter>
          <Shell />
        </BrowserRouter>
      </CatalogProvider>
    </ThemeProvider>
  );
}
