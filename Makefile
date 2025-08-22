include .env
export

.PHONY: help test test-unit test-integration test-api test-repositories test-e2e test-all pre-commit generate lint coverage_report coverage_integration cpu_profile mem_profile run docker-up docker-down docker-logs docker-clean docker-rebuild

help:
	cat Makefile

run:
	go run cmd/server/main.go

test: test-unit

test-unit:
	go test -v ./internal/... -tags=!integration

test-integration:
	go test -v ./test/integration/... -tags=integration

test-api:
	go test -v ./test/integration/api/... -tags=integration

test-repositories:
	go test -v ./test/integration/repositories/... -tags=integration

test-e2e:
	go test -v ./test/integration/e2e/... -tags=integration,e2e

test-all:
	go test -v ./internal/... -tags=!integration
	go test -v ./test/integration/... -tags=integration

generate:
	go generate ./...

lint:
	go fmt ./...
	find . -name '*.go' ! -path "./generated/*" -exec goimports -local simpleservicedesk/ -w {} +
	golangci-lint run ./...
	./check-go-generate.sh

coverage_report:
	go test -p=1 -coverpkg=./... -count=1 -coverprofile=.coverage.out ./internal/... -tags=!integration
	go tool cover -html .coverage.out -o .coverage.html
	open ./.coverage.html

coverage_integration:
	go test -p=1 -coverpkg=./... -count=1 -coverprofile=.coverage_integration.out ./test/integration/... -tags=integration
	go tool cover -html .coverage_integration.out -o .coverage_integration.html
	open ./.coverage_integration.html

cpu_profile:
	mkdir -p profiles
	go test -cpuprofile=profiles/cpu.prof -v ./test/integration/... -tags=integration
	go tool pprof -http=:6061 profiles/cpu.prof

mem_profile:
	mkdir -p profiles
	go test -memprofile=profiles/mem.prof -v ./test/integration/... -tags=integration
	go tool pprof -http=:6061 profiles/mem.prof

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

docker-rebuild:
	docker-compose down -v --remove-orphans
	docker-compose build --no-cache
	docker-compose up -d
