import React from "react";
import { NavLink } from "react-router-dom";
import { IconActivity, IconBook, IconGrid, IconKey, IconPlus, IconTrophy } from "../icons";

const NAV_ITEMS = [
  { to: "/", label: "Overview", icon: IconGrid, end: true },
  { to: "/create", label: "Create service", icon: IconPlus },
  { to: "/access", label: "Request access", icon: IconKey },
  { to: "/scorecards", label: "Scorecards", icon: IconTrophy },
  { to: "/observability", label: "Observability", icon: IconActivity },
  { to: "/docs", label: "Docs", icon: IconBook },
];

export function Sidebar({ open, onNavigate }: { open: boolean; onNavigate?: () => void }) {
  return (
    <nav id="app-sidebar" className={`app-sidebar ${open ? "open" : ""}`} aria-label="Primary">
      <div className="flex items-center gap-2 px-4 h-14 border-b border-pave-border">
        <img src="/brand/mark.svg" alt="" className="w-6 h-6" width={24} height={24} />
        <span className="wordmark text-sm font-semibold tracking-wide text-pave-text">
          Pavestack
        </span>
      </div>
      <div className="p-3 space-y-1">
        {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            onClick={onNavigate}
            className={({ isActive }) => `nav-link ${isActive ? "active" : ""}`}
          >
            <Icon className="w-4 h-4 shrink-0" />
            {label}
          </NavLink>
        ))}
      </div>
      <div className="mt-auto p-3 border-t border-pave-border text-[11px] text-pave-text-muted">
        <p>Pavestack Portal</p>
        <p className="mt-0.5">
          Catalog is generated from committed manifests. Write actions call pave-api.
        </p>
      </div>
    </nav>
  );
}
