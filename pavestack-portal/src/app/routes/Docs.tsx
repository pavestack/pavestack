import React, { useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { DOCS } from "../../lib/docsContent";
import { EmptyState } from "../components";
import { IconBook, IconSearch } from "../icons";

export function Docs() {
  const { section } = useParams<{ section?: string }>();
  const navigate = useNavigate();
  const [query, setQuery] = useState("");

  const activeId = section && DOCS.some((s) => s.id === section) ? section : DOCS[0].id;

  const filtered = useMemo(() => {
    if (!query.trim()) return DOCS;
    const lower = query.toLowerCase();
    return DOCS.map((s) => ({
      ...s,
      headings: s.headings.filter((h) => h.title.toLowerCase().includes(lower) || h.body.some((p) => p.toLowerCase().includes(lower))),
    })).filter((s) => s.headings.length > 0 || s.title.toLowerCase().includes(lower));
  }, [query]);

  const active = DOCS.find((s) => s.id === activeId) ?? DOCS[0];
  const activeFiltered = filtered.find((s) => s.id === activeId);
  const headingsToShow = query.trim() ? activeFiltered?.headings ?? [] : active.headings;

  return (
    <div className="grid gap-6 lg:grid-cols-[200px_1fr]">
      <aside>
        <label htmlFor="docs-search" className="sr-only">
          Search docs
        </label>
        <div className="relative mb-3">
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-pave-text-muted pointer-events-none">
            <IconSearch />
          </div>
          <input
            id="docs-search"
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search docs…"
            className="search-input pl-10"
          />
        </div>
        <nav aria-label="Docs sections" className="space-y-1">
          {(query.trim() ? filtered : DOCS).map((s) => (
            <button
              key={s.id}
              onClick={() => navigate(`/docs/${s.id}`)}
              className={`nav-link w-full text-left ${activeId === s.id ? "active" : ""}`}
            >
              {s.title}
            </button>
          ))}
        </nav>
      </aside>

      <div className="min-w-0">
        <h1 className="text-xl font-semibold text-pave-text mb-4">{active.title}</h1>

        {query.trim() && headingsToShow.length === 0 && (
          <EmptyState icon={<IconBook />} title="No matches" description={`No headings or paragraphs in "${active.title}" match "${query}".`} />
        )}

        <div className="space-y-6">
          {headingsToShow.map((h) => (
            <section key={h.id}>
              <h2 className="text-base font-semibold text-pave-text mb-2">{h.title}</h2>
              <div className="space-y-3">
                {h.body.map((p, i) => (
                  <p key={i} className="text-sm text-pave-text-secondary leading-relaxed">
                    {p}
                  </p>
                ))}
              </div>
            </section>
          ))}
        </div>
      </div>
    </div>
  );
}
