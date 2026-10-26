---
name: fix-frontend-bug
description: Diagnose and fix a frontend defect with the smallest regression-safe patch. Use for failed checks, regressions, or ambiguous Nuxt and Vue behavior.
---

# Fix Frontend Bug

Reproduce the failure and inspect the responsible code path before editing. Establish the root cause, apply the smallest complete patch at the owning layer, then rerun the failed check and the nearest affected gate. Report the root cause, changed scope, and regression proof.

Use the specialist skill that matches the problem when needed:

- Evidence-ranked root-cause analysis: `investigate-first`.
- Narrow fixes with regression proof: `surgical-patch`.
- Behavior-preserving cleanup required by the fix: `safe-refactor`.
- Visual, interaction, responsive, or accessibility defects: `ui-ux-pro-max`.
