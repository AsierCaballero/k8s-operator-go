# Contributing

Thanks for your interest in contributing to k8s-operator-go.

## How to contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Commit your changes using conventional commits
4. Push the branch (`git push origin feat/my-feature`)
5. Open a Pull Request

## Conventional commits

- `feat:` — new feature
- `fix:` — bug fix
- `test:` — tests
- `docs:` — documentation
- `refactor:` — code change without feature/fix
- `ci:` — CI/CD changes
- `chore:` — maintenance tasks

## Before submitting

- Run `make generate` to regenerate deepcopy and CRDs
- Run `make test` and ensure all tests pass
- Run `make lint` and fix any warnings
- Keep PRs focused on a single concern

## Code of conduct

Be respectful and constructive.
