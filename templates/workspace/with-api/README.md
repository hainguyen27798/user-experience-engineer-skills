# {{PROJECT_NAME}}

{{PROJECT_NAME}} is a small service-oriented workspace.

```text
.
├── {{UI_DIRECTORY}}/   # Nuxt application
├── {{API_DIRECTORY}}/  # FastAPI application managed by uv
├── .agents/            # Project-local skills and current framework references
├── .codex/             # Codex project agents
├── plan/               # Implementation plans
└── release.sh          # Reproducible release publisher
```

## Frontend

```bash
cd {{UI_DIRECTORY}}
npm run dev
```

## Backend

```bash
cd {{API_DIRECTORY}}
uv sync
uv run fastapi dev src/{{API_MODULE}}/main.py
```

The API exposes `GET /api/v1/health`. Integrate the frontend through an explicit HTTP contract rather than importing backend source.

## Checks

```bash
npm --prefix {{UI_DIRECTORY}} run lint
npm --prefix {{UI_DIRECTORY}} run test
npm --prefix {{UI_DIRECTORY}} run build
uv run --project {{API_DIRECTORY}} python -m compileall {{API_DIRECTORY}}/src
```

## Agent setup

Official Caveman, UI UX Pro Max, and Taste skills are installed into `.agents/skills/` from their default-branch HEAD and the resolved revisions are recorded in `skills-lock.json`. Project workflow skills live beside them, current Nuxt and Nuxt UI references live in `.agents/references/`, and Codex role definitions live in `.codex/agents/`.

Temporary output from Taste or any browser/design inspection belongs outside this repository. Keep only user-requested deliverables in documented service-owned paths.

## Release

Run `./release.sh` from a clean checkout. It builds the Nuxt server output and Python distributions, then publishes them under `frontend/` and `python/` on the orphan `release` branch without switching the source checkout. Release commits use `release YYYY/MM/DD HH:mm` in UTC.
