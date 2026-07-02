# Pavestack landing page

The marketing site for Pavestack. Plain static HTML/CSS/vanilla JS — no
build step, no framework, no dependencies.

## Structure

```
landing/
├── index.html          # the whole site (single page, section-anchored nav)
├── styles.css           # layout & components — consumes brand tokens, doesn't redefine colors
├── script.js             # theme toggle (localStorage + prefers-color-scheme) + mobile nav
├── assets/
│   └── brand/
│       ├── tokens.css     # copied from /brand/tokens.css — canonical color/type tokens
│       ├── mark.svg       # two-tone logo mark
│       ├── mark-mono.svg  # single-currentColor variant (used in the footer)
│       └── favicon.svg    # heavier-stroke mark, tuned for tab-size rendering
└── README.md
```

Diagrams (golden path pipeline, layered architecture) are hand-authored
inline `<svg>` in `index.html`, styled via CSS variables in `styles.css` so
they render correctly in both themes — there are no raster illustration
assets to keep in sync.

## Preview locally

Any static file server works. From the repo root:

```bash
npx serve landing
# or
python3 -m http.server --directory landing 8080
```

Then open the printed URL (e.g. `http://localhost:3000`).

## Deploy

Point any static host (S3 + CloudFront, Netlify, Vercel, GitHub Pages,
Cloudflare Pages, nginx, ...) at the `landing/` directory as the site root.
No build command or output directory configuration is needed — `index.html`
is the entry point and everything it references is a relative path inside
this directory.

When deployed alongside the self-service portal (`pavestack-portal/`) at a
sibling path, the nav's "Launch Portal" link and the footer's "Docs" link
(`../app/`, `../app/docs`) resolve correctly; adjust those two `href`s in
`index.html` if your deployment topology differs.

## Notes for future edits

- Color and type values come from `assets/brand/tokens.css`, which is a
  direct copy of `/brand/tokens.css` at the repo root. If the canonical
  tokens change, re-copy the file rather than hand-editing values here.
- Dark theme is the default (`:root`); light theme activates via
  `<html data-theme="light">`, toggled by `script.js` and persisted to
  `localStorage` under the `pavestack-theme` key.
- Keep new sections consistent with the existing type scale (`clamp()`
  based, defined per-component in `styles.css`) and avoid introducing a
  second color palette — everything should resolve through the CSS custom
  properties in `tokens.css`.
