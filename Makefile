.PHONY: test test-race vet tidy modules-check test-coverage generate check-generated verify-consumer security

test:
	go test ./...
	go -C docs test ./...
	python3 -m unittest discover -s scripts -p '*_test.py'

test-race:
	go test -race ./...
	go -C docs test -race ./...

vet:
	go vet ./...
	go -C docs vet ./...

tidy:
	go mod tidy
	go -C docs mod tidy

modules-check:
	./scripts/modules-check.sh

test-coverage:
	./scripts/test-unit.sh

generate:
	go -C docs run ./readme

check-generated:
	./scripts/check-generated.sh

verify-consumer:
	./scripts/verify-consumer.sh

security:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
	cd docs && go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
