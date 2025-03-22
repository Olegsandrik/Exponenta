FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/cache \
    go build -o exp ./cmd

FROM alpine:latest

COPY --from=builder /app/exp /exp

CMD ["/exp"]