# Generated Installation Manifest

This document lists what `create-ux-engineer` installs into every generated workspace. Transitive npm and Python dependencies are resolved by npm and uv and are recorded in their generated lockfiles.

## Required Host Tools

The initializer uses these existing tools; it does not install them:

| Tool | When required | Purpose |
| --- | --- | --- |
| `git` | Always | Preserve an existing repository or initialize a new one, and clone official skill sources. |
| `npm` | Always | Install and run the Nuxt frontend dependencies. |
| `uv` | Only with `--api` | Create and manage the FastAPI service. |

The initializer itself is a Go binary and never invokes `node`, `npx`, or a globally installed npm initializer. Frontend dependency installation still requires a working `npm` command and its supported Node.js runtime.

## Frontend Libraries

The generated `<project>-ui/package.json` receives these direct runtime dependencies from their current `latest` npm tags:

- `nuxt`
- `vue`
- `vue-router`
- `@nuxt/ui`

It also receives these direct development dependencies:

- `@nuxt/eslint@latest`
- `@vitejs/plugin-vue@latest`
- `@vue/test-utils@latest`
- `eslint@latest`
- `eslint-plugin-prettier@latest`
- `eslint-plugin-simple-import-sort@latest`
- `eslint-plugin-unused-imports@latest`
- `happy-dom@latest`
- `prettier@latest`
- `tailwindcss@latest`
- `typescript@^6.0.3`
- `vitest@latest`

npm records the exact resolved dependency graph in `<project>-ui/package-lock.json`.

## Optional Backend Library

When `--api` is selected, uv creates `<project>-api` as a packaged application and adds the direct dependency `fastapi[standard]`. uv records the resolved Python dependency graph in `<project>-api/uv.lock`.

## Generated Agent Instructions

Every workspace receives:

- `AGENTS.md` at the workspace root for cross-service conventions.
- `<project>-ui/AGENTS.md` for Nuxt ownership, commands, visual consistency, and Taste artifact rules.
- `<project>-api/AGENTS.md` when `--api` is selected, for FastAPI ownership and commands.

## Codex Agents

The following profiles are written to `.codex/agents/` and registered by `.codex/config.toml`:

| Agent | Responsibility |
| --- | --- |
| `planner` | Create repository-grounded implementation plans without editing unless explicitly asked. |
| `reviewer` | Review correctness, security, regressions, and missing tests. |
| `cavecrew-investigator` | Locate relevant code and return compressed evidence. |
| `cavecrew-builder` | Implement a narrowly assigned one- or two-file change. |
| `cavecrew-reviewer` | Return concise, actionable findings for an assigned diff. |

## Official Skills

The initializer shallow-clones the current default-branch HEAD of each official repository and copies the listed skill directories without modifying them. The resolved commit and installed-directory SHA-256 are recorded in `skills-lock.json`.

### Caveman

Source: [`JuliusBrussee/caveman`](https://github.com/JuliusBrussee/caveman)

- `caveman`
- `ultracave`
- `megacave`
- `cavecrew`
- `caveman-explore`
- `caveman-compress`
- `lean-build`
- `investigate-first`
- `surgical-patch`
- `safe-refactor`
- `migration`
- `caveman-review`
- `verify-and-stop`
- `caveman-commit`
- `caveman-learn`

### UI UX Pro Max

Source: [`nextlevelbuilder/ui-ux-pro-max-skill`](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill)

- `ui-ux-pro-max`

### Taste

Source: [`Leonxlnx/taste-skill`](https://github.com/Leonxlnx/taste-skill)

- `design-taste-frontend`, copied from the upstream `skills/taste-skill` directory.

Taste screenshots, browser profiles, reports, caches, previews, and history must remain in an OS temporary directory outside the generated repository unless the user explicitly requests a particular deliverable.

## Project Workflow Skills

These repository-owned orchestration skills are embedded in the binary and written to `.agents/skills/`:

| Skill | Responsibility |
| --- | --- |
| `create-implementation-plan` | Create deterministic frontend or backend implementation plans. |
| `design-frontend-ui` | Derive visual decisions from existing project evidence without inventing a style. |
| `implement-frontend-change` | Implement planned Nuxt changes at the responsible layer. |
| `review-frontend-change` | Review correctness, accessibility, scope, and visual consistency. |
| `test-frontend-change` | Select and run task-relevant Nuxt checks. |
| `verify-frontend-change` | Prove acceptance conditions without expanding scope. |
| `fix-frontend-bug` | Diagnose and repair frontend defects with regression proof. |

## Downloaded References

The initializer stores the current official reference files in `.agents/references/`:

- `nuxt-llms-full.txt` from `https://nuxt.com/llms-full.txt`
- `nuxt-ui-llms-full.txt` from `https://ui.nuxt.com/llms-full.txt`
