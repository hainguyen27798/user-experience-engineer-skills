# Nuxt Frontend

## Development

```bash
npm install
npm run dev
```

Run `npm run lint` for static checks, `npm run test` for unit tests in happy-dom, and `npm run build` for routing, module, configuration, or SSR changes.

Browser API requests should use `runtimeConfig.public.apiBase`, which defaults to `/api`. Copy `.env.example` to `.env` to target a separate local API through `NUXT_PUBLIC_API_BASE`; this setting does not configure a Nuxt proxy.

## Folder conventions

```text
app/
├── pages/       # File-based route entry points
├── components/  # Reusable and feature-owned UI components
├── composables/ # Reusable stateful Vue logic and API orchestration
├── layouts/     # Shared page shells
└── assets/      # Source styles and bundled assets
public/          # Files served unchanged at stable URLs
server/          # Nuxt-only server routes and server adapters
tests/           # Cross-feature or browser-level suites only
```

Keep focused unit tests beside their owner as `*.test.ts`. Create a directory with its first real file instead of adding an empty placeholder tree. See `AGENTS.md` for ownership and workspace integration rules.
