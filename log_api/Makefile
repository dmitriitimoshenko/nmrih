test-go:
	go build main.go
	go test ./...

docker-image-build:
	docker build -t api:latest .

docker-run:
	docker run -d api:latest

docker-stop-and-remove:
	docker ps -q | xargs -r docker stop && \
    docker ps -aq | xargs -r docker rm

docker-clean-up:
	docker images -aq | xargs -r docker rmi

go-mod:
	go mod vendor
	go mod tidy

test-with-coverage:
	go test -parallel=1 -count=1 ./... -coverprofile cover.out
