---
name: test-frontend-change
description: Select and run focused tests for a Nuxt or Vue change, then report reproducible evidence.
---

# Test Frontend Change

Derive tests from the requested behavior, changed code, and material risks. Run the narrowest relevant checks first. Add or change tests only when behavior needs regression coverage.

Inspect `{{UI_DIRECTORY}}/package.json` before selecting checks. Run frontend package commands from the repository root with `npm --prefix {{UI_DIRECTORY}}`. Run lint and test scripts when they exist and cover the change; report unavailable checks rather than assuming scripts exist or adding tooling solely for verification. Run `npm --prefix {{UI_DIRECTORY}} run build` for routing, configuration, module, or SSR changes. Distinguish pass, fail, unavailable, and blocked, and report the exact commands and results.
