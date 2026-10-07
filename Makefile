.PHONY: generate build test dev clean

generate:
	go tool templ generate
	go tool sqlc generate

build: generate
	go build -o cockadeck ./cmd/cockadeck

test: generate
	go test ./...

dev: generate
	COCKADECK_JWT_SECRET=dev-secret COCKADECK_ENV=dev go run ./cmd/cockadeck

clean:
	rm -f cockadeck cockadeck.db cockadeck.db-*
