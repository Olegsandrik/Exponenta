FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o exp ./cmd

FROM alpine:latest

COPY --from=builder /app/exp /exp

CMD ["/exp"]