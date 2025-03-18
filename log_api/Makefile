# Go commands

go-mod:
	go mod vendor
	go mod tidy

test:
	go build main.go
	go test ./...

test-with-coverage:
	go test -parallel=1 -count=1 ./... -coverprofile cover.out

format:
	golines --base-formatter="gofmt" -w -m 120 --ignored-dirs="vendor tmp data" --ignore-generated .
	gofumpt -w .

lint:
	golangci-lint run -c=golangci.yaml

# Docker commands

docker-image-build:
	docker build -t log_api:latest .
