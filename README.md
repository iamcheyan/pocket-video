# pocket-video

PocketVideo — a Bilibili favorite-folder offline archiver built as a PocketJS
reference application: TypeScript/React client, Go service, SQLite, and yt-dlp
behind a provider adapter.

- Application and documentation: [`apps/pocket-video/README.md`](apps/pocket-video/README.md)
- Architecture: [`apps/pocket-video/docs/ARCHITECTURE.md`](apps/pocket-video/docs/ARCHITECTURE.md)
- API: [`apps/pocket-video/docs/API.md`](apps/pocket-video/docs/API.md)
- Screenshots: [`apps/pocket-video/docs/screenshots/`](apps/pocket-video/docs/screenshots/)

```sh
make backend      # Go service, mock mode, http://127.0.0.1:8080
make frontend     # Vite client, http://127.0.0.1:5173
make test         # Go + frontend tests, production build, screenshot verification
```
