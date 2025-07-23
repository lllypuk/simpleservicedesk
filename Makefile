include .env
export

.PHONY: help unit_test integration_test test generate lint coverage_report cpu_profile mem_profile

help:
	cat Makefile

unit_test:
	go test -v ./internal/...

integration_test:
	go test -v ./integration_test/...

test: unit_test integration_test

generate:
	go generate ./...

lint:
	go fmt ./...
	find . -name '*.go' ! -path "./generated/*" -exec goimports -local simpleservicedesk/ -w {} +
	golangci-lint run ./...
	./check-go-generate.sh

coverage_report:
	go test -p=1 -coverpkg=./... -count=1 -coverprofile=.coverage.out ./...
	go tool cover -html .coverage.out -o .coverage.html
	open ./.coverage.html

cpu_profile:
	go test -cpuprofile=profiles/cpu.prof  ./e2e_test
	go tool pprof -http=:6061 profiles/cpu.prof

mem_profile:
	go test -memprofile=profiles/mem.prof ./e2e_test
	go tool pprof -http=:6061 profiles/mem.prof