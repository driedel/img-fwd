---
name: commit-style
description: Use when writing a git commit message for this repo. Enforces Conventional Commits format in English with the types and scopes used here.
---

# Commit Style

## Language

All commit messages must be in **English**.

## Format

Use [Conventional Commits](https://www.conventionalcommits.org/) with optional scope:

```
<type>[(<scope>)]: <description>

[optional body]

[optional footer(s)]
```

## Types used in this repo

| Type | Use when |
|---|---|
| `feat` | New behavior or capability |
| `fix` | Bug fix |
| `test` | Adding or updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `docs` | README, AGENTS.md, comments |
| `chore` | Build, CI, Docker, dependency updates |
| `perf` | Performance improvement |

## Scopes (optional)

| Scope | Area |
|---|---|
| `app` | Go source code |
| `proxy` | Request handling / routing logic |
| `docker` | Dockerfile, entrypoint, compose files |
| `ci` | GitHub Actions |
| `readme` | README.md |
| `agents` | AGENTS.md, OpenCode config |

## Examples

```
feat(proxy): add blur transformation support

fix(app): correct rs height parsing when only width provided

test(app): add origin forwarding test for non-imgproxy params

refactor(proxy): extract format detection into autoFormat function

docs(readme): document automatic format conversion table

chore(ci): bump Go version to 1.22 in GitHub Actions

docker: pin imgproxy base image to v3.24
```

## Rules

- Keep the description line under 72 characters.
- Use imperative mood: "add" not "added", "fix" not "fixed".
- If the commit closes an issue, add `Closes #<number>` in the footer.
- Do not include issue numbers in the subject line.
