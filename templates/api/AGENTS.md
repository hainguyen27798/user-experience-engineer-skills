# FastAPI Service Instructions

## Folder Conventions

- `src/{{API_MODULE}}/main.py` is the composition root. Configure FastAPI and include routers here; do not put feature logic in it.
- `src/{{API_MODULE}}/api/` owns HTTP routers, request/response schemas, and transport-level dependencies.
- `src/{{API_MODULE}}/core/` owns cross-cutting configuration, errors, security primitives, and application lifecycle code.
- `src/{{API_MODULE}}/modules/<feature>/` owns feature-specific domain rules and use cases. Keep dependencies pointing inward toward the feature rather than toward HTTP handlers.
- `src/{{API_MODULE}}/infrastructure/` owns database, email, queue, and external-service adapters.
- `tests/` mirrors service behavior. Name tests `test_*.py` and keep fixtures close to the narrowest shared scope.

Create a directory only when its first owned file is needed; do not add empty scaffolding or placeholder modules.

## Working Agreements

- Keep route handlers thin and move reusable behavior to its owning module.
- Validate untrusted input at the HTTP boundary and enforce authorization again at the domain operation that owns it.
- Keep secrets in ignored environment files and commit only safe examples.
- Integrate with the frontend through an explicit HTTP contract; never import frontend implementation code.
- Add focused tests for behavior changes and run the smallest relevant checks before broader service checks.

## Commands

- `uv sync` — install locked dependencies.
- `uv run fastapi dev src/{{API_MODULE}}/main.py` — start the development API.
- `uv run python -m compileall src` — compile-check the service when no narrower test exists.
