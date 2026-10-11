.PHONY: gen lint tidy fmt go-mod-update test-unit test-integration test-component


test-unit:
	go test -race ./...

lint:
	go tool golangci-lint run ./...

fmt:
	go tool gofumpt -l -w .

go-mod-update:
	go get -u ./...

tidy:
	go mod tidy
