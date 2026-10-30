# Tests

Two suites protect the app:

- `backend/*_test.go` — Go unit and integration tests for migrations, the mock
  and yt-dlp adapters, the REST API, pause/resume, queue cleanup, settings
  validation, and automatic sync.
- `tests/*.test.ts` — frontend tests for the API client, routing, formatting,
  theme resolution, and i18n parity. They run on Node's built-in test runner
  with type stripping, so no extra test framework is required.

Run everything from the repository root:

```sh
make test
```

`make test` also builds the production bundle and verifies the screenshot
gallery (PNG dimensions, unique hashes, and no horizontal overflow at 320×240
through 1280×720).

Run one side alone:

```sh
cd apps/pocket-video/backend && go test ./...
cd apps/pocket-video && node --test --experimental-strip-types tests/*.test.ts
```
