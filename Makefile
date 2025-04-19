lint:
	golangci-lint run

fmt:
	golangci-lint fmt

mocks:
	mockery --config .mockery.yaml

unit-test:
	go test -count=1 -cover -v $(shell go list ./... | grep -v '/e2e$$')

e2e-test:
	go test -count=1 -cover -v github.com/lovelaze/nebula-sync/e2e

test: unit-test e2e-test
