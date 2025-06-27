# Contributing Guide

## Branching Strategy
- **main**: Production-ready code only.
- **develop**: Integration branch for all features and fixes.
- **feature/<name>**: New features (e.g., `feature/user-auth`).
- **hotfix/<name>**: Critical fixes for production (e.g., `hotfix/login-bug`).
- **bugfix/<name>**: Non-critical bug fixes (e.g., `bugfix/typo`).
- **chore/<name>**: Maintenance, tooling, or non-feature changes.
- **refactor/<name>**: Code refactoring without feature changes.

## Pull Requests
- PRs must target `develop` unless hotfix (then target `main`).
- Use the PR template. Fill all sections.
- Assign reviewers as per CODEOWNERS.
- Reference related issues (e.g., `Fixes #123`).
- Ensure all status checks pass (test, build, lint, security).
- At least 2 approvals required.

## Commits
- Use clear, concise commit messages (imperative mood).
- Example: `fix: handle nil pointer in user service`
- Squash commits before merging if possible.

## Issues
- Use issue templates for bugs/features.
- Provide clear reproduction steps and context.

## Code Style & Quality
- Pre-commit hooks and linters must pass locally.
- Write unit tests for new features and bug fixes.
- Document public functions and exported types.

## Security
- Never commit secrets or credentials.
- Use GitHub Secrets for sensitive data.
- Run security and dependency scans on all PRs.

## Access
- Only tech team members are collaborators.
- All code changes require PR review and approval.

## Questions
- For access or process issues, tag @kaushik in an issue or PR.
