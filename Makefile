.PHONY: build test vet check coverage clean

build:
	go build -o mtb .

test:
	go test ./... -count=1

vet:
	go vet ./...

check: vet build test

coverage:
	go test ./internal/tools/ -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

clean:
	rm -f mtb coverage.out coverage.html
