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

# Cross-compile for the VPS and ship the binary. Usage:
#   make deploy DEPLOY_HOST=user@vps.example.com
DEPLOY_HOST ?=

deploy:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cockadeck-linux ./cmd/cockadeck
	scp cockadeck-linux $(DEPLOY_HOST):/tmp/cockadeck
	ssh $(DEPLOY_HOST) 'sudo install -m755 /tmp/cockadeck /usr/local/bin/cockadeck && sudo systemctl restart cockadeck'

# First-time provisioning: ship the binary + deploy/ files, then on the host:
#   make deploy-bundle DEPLOY_HOST=root@<vps>
#   ssh $(DEPLOY_HOST) 'DOMAIN=cockadeck.example.com sh /tmp/cockadeck-deploy/bootstrap.sh'
deploy-bundle:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cockadeck-linux ./cmd/cockadeck
	ssh $(DEPLOY_HOST) 'rm -rf /tmp/cockadeck-deploy && mkdir -p /tmp/cockadeck-deploy'
	scp cockadeck-linux $(DEPLOY_HOST):/tmp/cockadeck
	scp deploy/* $(DEPLOY_HOST):/tmp/cockadeck-deploy/

clean:
	rm -f cockadeck cockadeck-linux cockadeck.db cockadeck.db-*
