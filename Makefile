.PHONY : generate build check build-test deploy

generate:
	cd internal/database && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
build:
	go build -o bin/ ./cmd/...
test:
	go test ./...
