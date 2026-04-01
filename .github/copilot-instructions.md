# GitHub Copilot Instructions

## Commit & Push Rules

**NEVER run `git commit` or `git push` without explicit approval from the user.**

Before committing or pushing:
1. Show the user exactly what will be staged/committed (file list + summary of changes).
2. Wait for the user to respond with an explicit "yes" or equivalent confirmation.
3. Only then proceed with the git operation.

This rule applies to every session, regardless of how minor the change appears.

## Branching

- Active development branch: `dev/ci-cd`
- Production branch: `main` (protected — CI/CD deploys from here)
- Never push directly to `main`; changes go through PRs from `dev/ci-cd`

## Go Workspace

- The repo uses `go.work` at the root (workspace mode).
- Services: `api_gateway`, `auth_service`, `email_service`, `order_service`, `product_service`, `user_client`, `user_service`
- When adding a new dependency to a service, run `go get` inside that service's directory and commit the updated `go.mod` + `go.sum`.
- Never add `replace` directives to individual `go.mod` files — cross-service references are handled by `go.work`.

## CI/CD Conventions

- Workflows live in `.github/workflows/`.
- Each service has its own deploy workflow (`<service>-deploy.yml`).
- Workflow structure: Checkout → AWS creds (push only) → ECR login (push only) → Set up Go → Run tests → Build image (push only) → Trivy scan (push only) → Push ECR (push only) → K8s deploy (push only) → Verify (push only).
- The `pull_request` trigger runs only Steps 1 (checkout), 4 (set up Go), and 5 (tests).
- Deploy steps are guarded with `if: github.event_name == 'push'`.
