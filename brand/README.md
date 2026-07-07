# Pavestack brand assets

Canonical source of truth for the Pavestack identity. `landing/` and
`pavestack-portal/` both consume these files directly (copied into their
public/asset directories at build time) — do not fork colors or the mark
into a second definition.

## Mark

`mark.svg` — hexagon ("paver tile") + forward chevron ("pipeline
progression"). Two-tone: `#14B8A6` (teal) hexagon stroke, `#F59E0B` (amber)
chevron. Verified legible from 24px (nav bar / favicon tab) through 200px+
(hero lockups). Do not add gradients, drop shadows, or 3D bevels to it.

`mark-mono.svg` — single-`currentColor` variant for contexts where only one
ink color is available (e.g. inside a badge or footer rule).

`favicon.svg` — same geometry, heavier stroke weights, tuned for
sub-32px rendering in browser tabs.

## Why this replaces `logo.svg` / `logo.png` / `assets/banner.svg`

The original assets were an isometric stacked-diamond mark in blue→purple
gradients — the generic "AI/SaaS" look the 2026 landing-page brief
explicitly calls out to avoid. They are left in place for git history but
are no longer referenced from the landing page, portal, or README; treat
`brand/` as canonical going forward. See `.agents/memory/decisions.md`.

## Color tokens

See `tokens.css` for the full CSS custom-property set (dark default +
`[data-theme="light"]` override). Do not introduce a third palette
(e.g. a Tailwind default indigo/violet) anywhere in landing or portal code —
all UI color must resolve through these tokens so the two surfaces stay
visually part of one product.

## Type

- Display / wordmark: **Space Grotesk** (geometric, technical — used for
  the landing page headings and the nav wordmark). Loaded via Google Fonts
  with `<link rel="preconnect">` + `font-display: swap`.
- Body / UI: **Inter**.
- Code / tabular data: **JetBrains Mono** — also used for all numeric
  table columns (`font-variant-numeric: tabular-nums`).

The portal caps display usage: Space Grotesk appears only in the nav
wordmark, never for in-app headings (portal heading scale tops out at
`--text-xl`, per INTENT_SPEC). The landing page uses Space Grotesk more
theatrically (hero, section titles) — that asymmetry is intentional, see
`AGENTS.md`.
