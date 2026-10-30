.PHONY: test backend frontend dev

test:
	cd apps/pocket-video/backend && go test ./...
	cd apps/pocket-video && node tests/check-i18n.mjs && npm --prefix frontend run build

backend:
	cd apps/pocket-video/backend && go run .

frontend:
	cd apps/pocket-video/frontend && npm run dev

dev:
	@echo 'Run make backend and make frontend in separate terminals.'

screenshots:
	./scripts/screenshots.sh

