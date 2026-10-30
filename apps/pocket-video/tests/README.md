# Integration tests

Backend behavior and SQLite are tested with `go test ./...` in `backend`. The frontend build performs TypeScript checking; `scripts/check-i18n.mjs` checks that all three locale dictionaries expose the same keys.
