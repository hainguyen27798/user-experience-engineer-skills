---
name: review-frontend-change
description: Review a frontend diff for correctness, regressions, accessibility, scope, and maintainability.
---

# Review Frontend Change

Inspect the user request, task diff, and relevant code paths in `{{UI_DIRECTORY}}/`. For visual changes, read `{{UI_DIRECTORY}}/docs/design-system.md` and compare against the closest existing screens, layouts, components, and tokens. Flag invented layout, color, font, type-size, spacing, sizing, radius, border, shadow, responsive, or interaction rules that lack project evidence or explicit user approval. Also check accessibility, Nuxt SSR behavior, regression risk, ownership, and scope drift. Report actionable findings with precise locations. Do not edit unless fixes were requested.

Use the specialist skill that matches the review when needed:

- Concise code and diff review: `caveman-review`.
- Delegated diff review when explicitly requested: `cavecrew`.
- UI/UX, accessibility, and responsive review: `ui-ux-pro-max`.

If Taste is used for visual review, put screenshots, reports, browser profiles, generated previews, caches, and history in an OS temporary directory outside the repository and leave no tool artifacts behind.
