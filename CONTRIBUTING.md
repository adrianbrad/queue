# Contributing

Thanks for your interest in improving `queue`. This document is the contract for contributions.

## Getting started

```bash
git clone https://github.com/adrianbrad/queue.git
cd queue
go test ./...
```

Go 1.20+ is required.

## Reporting bugs and proposing features

- Bugs: <https://github.com/adrianbrad/queue/issues>
- Security issues: see [SECURITY.md](SECURITY.md) — do **not** file public issues for security problems.
- Feature ideas: open an issue first to discuss scope before opening a PR.

## Pull request workflow

1. Fork, then branch from `main`. Use [conventional-branch](https://conventional-branch.github.io/) prefixes: `feature/`, `bugfix/`, `chore/`, `hotfix/`.
2. Make changes on the branch. Keep PRs focused — one logical change per PR.
3. Ensure every local gate passes before pushing:
   ```bash
   make lint          # golangci-lint --fix; must be clean
   make test-coverage # unit + race tests; fails if coverage != 100%
   ```
4. Open a PR against `main`.
5. CI must be green: `Lint`, `Test`, `Analyze` (CodeQL), `gitleaks`, `grype`.

## Coding standards

- Style: `gofumpt` + `goimports` (applied by `make lint --fix`).
- Lint config: [`.golangci.yml`](.golangci.yml). New code must satisfy every enabled linter.
- Tests: every exported behavior must be covered. Coverage gate is **100%** of statements; PRs that drop coverage will fail CI.
- Concurrency: queue implementations must remain safe under `-race`. Add a regression test when touching locking.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <summary>

<optional body explaining the *why*>
```

Types used here: `feat`, `fix`, `perf`, `refactor`, `test`, `docs`, `style`, `build`, `chore`.

## Review

- A CODEOWNER approval is required (see [`.github/CODEOWNERS`](.github/CODEOWNERS)).
- Status checks must pass; the branch must be up to date with `main`.
- Discussions are resolved before merge; commits are squash-merged.

## Release process

Tagged releases follow [SemVer](https://semver.org/). Maintainers bump the version by tagging `vX.Y.Z` on `main`; `CHANGELOG.md` is updated in the same commit.
