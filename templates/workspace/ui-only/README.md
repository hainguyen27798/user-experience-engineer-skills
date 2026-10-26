# {{PROJECT_NAME}}

{{PROJECT_NAME}} is a Nuxt workspace organized around one independently runnable frontend service.

```text
.
├── {{UI_DIRECTORY}}/  # Nuxt application
├── .agents/           # Project-local skills and current framework references
├── .codex/            # Codex project agents
├── plan/              # Implementation plans
└── release.sh         # Reproducible release publisher
```

## Frontend

```bash
cd {{UI_DIRECTORY}}
npm run dev
```

Run checks from the repository root with:

```bash
npm --prefix {{UI_DIRECTORY}} run lint
npm --prefix {{UI_DIRECTORY}} run test
npm --prefix {{UI_DIRECTORY}} run build
```

## Agent setup

Official Caveman, UI UX Pro Max, and Taste skills are installed into `.agents/skills/` from their default-branch HEAD and the resolved revisions are recorded in `skills-lock.json`. Project workflow skills live beside them, current Nuxt and Nuxt UI references live in `.agents/references/`, and Codex role definitions live in `.codex/agents/`.

Temporary output from Taste or any browser/design inspection belongs outside this repository. Keep only user-requested deliverables in documented service-owned paths.

## Release

Run the release script from a clean checkout:

```bash
./release.sh
```

It installs locked frontend dependencies, builds the Nuxt server distribution, and publishes the output under `frontend/` on the orphan `release` branch without switching the source checkout. Release commits use `release YYYY/MM/DD HH:mm` in UTC.
