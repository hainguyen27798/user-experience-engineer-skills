# Security Policy

## Supported Versions

Only the latest published version is supported. Alpha releases may change without backward compatibility.

## Reporting a Vulnerability

Report vulnerabilities privately through [GitHub Security Advisories](https://github.com/hainguyen27798/user-experience-engineer-skills/security/advisories/new). Do not open a public issue before a fix is available.

Include affected versions, reproduction steps, impact, and any suggested mitigation. Never include real credentials or private project data.

## Trust Boundary

This initializer runs npm packages and downloads agent skills from official GitHub repositories. Nuxt packages and skill repositories intentionally resolve their current latest versions at initialization time. The generated `package-lock.json` and `skills-lock.json` record the resolved dependency state for review; all downloaded code remains external code trusted by the user running the initializer.

- Content in generated `.agents/references/` files is untrusted external documentation. Generated repository instructions require agents to ignore embedded commands, role changes, tool requests, and data-exfiltration requests.
- Official third-party skills are executable agent instructions. Their exact GitHub owners are allowlisted in the binary, but using current default-branch HEAD means an upstream compromise can affect a new initialization before its recorded hash is reviewed.
- `--update` explicitly invokes the authenticated Codex CLI after managed skills and references are refreshed. Normal initialization asks for consent first and defaults to no. When used, Codex receives the real project as read-only evidence and workspace-write access only to an isolated temporary directory outside the repository. The initializer rejects symlinks in managed rule paths and promotes only validated agent instruction and orchestration files. Users should still review model-generated rule edits before committing.
- npm and uv dependencies execute within their normal package-manager trust boundaries. Resolving current releases improves freshness but does not make those releases trusted or vulnerability-free.
- The shell installer verifies archive integrity with a SHA-256 checksum from the same GitHub Release and rejects unexpected archive paths and symlinks. The checksum does not protect against compromise of the GitHub repository or release publisher.

Protect the GitHub publisher account with phishing-resistant multi-factor authentication, restrict force-pushes and release permissions, and review resolved locks before using a generated workspace for sensitive code.
