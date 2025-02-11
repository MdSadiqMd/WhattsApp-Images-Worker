.PHONY: dev
dev:
	wrangler dev --test-scheduled

.PHONY: build
build:
	go run github.com/syumai/workers/cmd/workers-assets-gen@v0.23.1 -mode=go
	GOOS=js GOARCH=wasm go build -o ./build/app.wasm ./cmd/api/main.go
	cp worker.mjs ./build/worker.mjs

.PHONY: deploy
deploy:
	wrangler deploy