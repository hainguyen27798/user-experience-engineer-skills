---
name: design-frontend-ui
description: Design a frontend screen, component, or interaction. Use for visual, responsive, interaction, or accessibility decisions.
---

# Design Frontend UI

Inspect the closest existing screens, layouts, tokens, components, and visual patterns in `{{UI_DIRECTORY}}/`. Read `{{UI_DIRECTORY}}/docs/design-system.md` if it exists. Existing implementation evidence is authoritative unless the user explicitly requests a visual-system change.

Define concrete implementation decisions by reusing evidenced layout, color roles, typography, type sizes, spacing, sizing, radius, borders, shadows, interaction states, responsive behavior, and accessibility patterns. Do not invent missing visual rules or let a generic design-skill recommendation override the project. If evidence is absent, conflicting, or materially ambiguous, stop and ask the user. Promote explicitly approved reusable decisions into the design system; keep request-specific exceptions local to the requested UI. Produce design decisions, not production markup.

Use `ui-ux-pro-max` only for guidance or checks that are mapped back to existing project evidence. It does not authorize a new style direction.

When using `design-taste-frontend`, direct screenshots, reports, browser profiles, generated previews, caches, and history to an OS temporary directory outside the repository. Do not keep Taste artifacts unless the user explicitly requested that exact file as a deliverable.
