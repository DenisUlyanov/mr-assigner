FROM golang:1.15 AS builder

# enable Go modules support
ENV GO111MODULE=on

WORKDIR /app

COPY . .

# install go mod, compiler protobuf
RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -v -o go_service cmd/goapp/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go_service /usr/local/bin
CMD ["go_service", "service", "http"]
