# FastAPI service

## Development

```bash
uv sync
uv run fastapi dev src/{{API_MODULE}}/main.py
uv run ruff check .
uv run ruff format --check .
uv run mypy src
```

The development API is available at `http://localhost:8000`; health is available at `GET /api/v1/health`.

Keep route handlers, domain logic, dependencies, tests, and environment examples inside this service. Expose behavior to the frontend through an explicit HTTP contract.

## Folder conventions

```text
src/{{API_MODULE}}/
├── main.py          # Composition root
├── api/             # HTTP routers and transport schemas
├── core/            # Config, errors, security, lifecycle
├── modules/         # Feature-owned domain rules and use cases
└── infrastructure/  # Database and external-service adapters
tests/               # Service behavior tests
```

Create each directory with its first real file; do not add empty placeholder scaffolding. See `AGENTS.md` for ownership rules.

Use `uv add <package>` and `uv remove <package>` for runtime dependencies. Use their `--dev` forms for development tools. Commit `pyproject.toml` and `uv.lock` together; never edit the lockfile manually.
