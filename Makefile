.PHONY: test backend frontend dev screenshots verify

test:
	cd apps/pocket-video/backend && go test ./...
	cd apps/pocket-video && node --test --experimental-strip-types tests/*.test.ts
	cd apps/pocket-video/frontend && npm run build
	python3 scripts/capture-screenshots.py --verify

backend:
	cd apps/pocket-video/backend && go run .

frontend:
	cd apps/pocket-video/frontend && npm run dev

dev:
	@echo 'Run make backend and make frontend in separate terminals.'

screenshots:
	python3 scripts/capture-screenshots.py

verify:
	python3 scripts/capture-screenshots.py --verify
