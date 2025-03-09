FROM golang:1.24

WORKDIR /app

COPY . /app

RUN make go-mod
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /log_parser

CMD ["/log_parser"]
