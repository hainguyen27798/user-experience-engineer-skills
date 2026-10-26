# Frontend Project Instructions

## Stack

This project uses the latest stable Nuxt, Vue, Nuxt UI, Tailwind CSS, and the Nuxt ESLint module, with a compatible TypeScript release. Project-local skills are installed at the workspace root in `../.agents/skills/`; load them only when their descriptions match the task.

For every non-trivial frontend change, load `../.agents/skills/create-implementation-plan/SKILL.md` before editing. Plans live in `../plan/`. Then load only the stage skill needed for design, implementation, review, testing, verification, or bug fixing. Keep the plan current when evidence, scope, or checks change.

## Folder Conventions

- `app/pages/` owns file-based route entry points. Keep pages thin and compose feature components.
- `app/components/` owns reusable presentation and interaction components. Group feature-only components by feature instead of creating a global grab bag.
- `app/composables/` owns reusable stateful Vue logic and API client orchestration.
- `app/layouts/` owns shared page shells; `app/assets/` owns source styles and bundled assets.
- `public/` owns files served unchanged at stable URLs.
- `server/` owns Nuxt server routes and server-only adapters. When the workspace has a FastAPI service, keep domain rules, persistence, authentication, and cross-client APIs in that service rather than duplicating them here.
- Keep focused unit tests beside their owner as `*.test.ts`; use `tests/` only for cross-feature or browser-level suites.

## Working Agreements

- Reuse Nuxt UI components and semantic design tokens before creating custom primitives.
- Treat `docs/design-system.md` as the project-specific visual source of truth. Read it for every visual change; establish or update its decisions in the design stage before implementation.
- Before implementing UI, inspect the closest existing screens, layouts, components, and tokens. Match their composition, color roles, typography, type sizes, spacing, sizing, radius, borders, shadows, responsive behavior, and interaction patterns. Do not invent a new visual language or import a generic design-skill style. If no established pattern answers a material visual choice, stop and ask the user unless they explicitly requested a new direction.
- Do not add one-off colors, fonts, spacing, radius, or shadows when an existing token fits. Keep raw visual values in the theme/token layer and record new reusable values in the design system.
- Promote explicitly approved, reusable visual feedback into `docs/design-system.md`. Keep task-specific exceptions in the active plan; do not infer global preferences from a single correction.
- Keep route components in `app/pages/`, reusable UI in `app/components/`, stateful logic in `app/composables/`, and server behavior in `server/`.
- Preserve SSR behavior; do not access browser-only globals outside client-safe lifecycle hooks.
- Use semantic HTML, keyboard-accessible interactions, visible focus states, and sufficient contrast.
- Run `npm run lint` after code changes, `npm run test` for tested frontend behavior, and `npm run build` for changes that affect routing, configuration, modules, or SSR.
- Keep secrets in local environment files. Only expose values prefixed with `NUXT_PUBLIC_` when they are safe for browsers.
- Taste and browser/design inspection tools must keep screenshots, reports, profiles, previews, caches, and history in an OS temporary directory outside the repository unless the user explicitly requests a deliverable.

## Commands

- `npm run dev` — start local development.
- `npm run lint` — check JavaScript, TypeScript, and Vue files.
- `npm run lint:fix` — apply safe ESLint fixes.
- `npm run test` — run unit tests in happy-dom.
- `npm run build` — produce a production build.
