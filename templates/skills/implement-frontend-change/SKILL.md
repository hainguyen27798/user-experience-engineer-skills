---
name: implement-frontend-change
description: Implement a planned Nuxt or Vue frontend feature, refactor, or migration at the narrowest responsible layer. Use only after an implementation plan is coherent.
---

# Implement Frontend Change

Read the active plan and `{{UI_DIRECTORY}}/docs/design-system.md`, then inspect the closest existing screen and component patterns before coding. Match evidenced layout, color roles, fonts, type sizes, spacing, sizing, radius, borders, shadows, responsive behavior, and interactions. Do not invent a visual rule, introduce a generic design-skill aesthetic, or create a one-off value when an existing pattern or token fits. If the visual contract does not answer a material choice, return to design; if repository evidence still does not answer it, ask the user. Implement the next step in `{{UI_DIRECTORY}}/` with existing Nuxt UI components, SSR, and accessibility intact. Mark the step complete only after its local condition passes. If evidence changes scope, update the plan before continuing.
