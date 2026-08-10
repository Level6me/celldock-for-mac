# Multi-arch Dockerfile for CellDock Web (Raspberry Pi / x86 Linux)
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .

RUN go build -o celldock-web main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata alsa-utils

WORKDIR /app
COPY --from=builder /app/celldock-web /app/celldock-web

EXPOSE 8080 1080

ENV LISTEN_ADDR=:8080
ENV DATA_DIR=/app/data

CMD ["/app/celldock-web"]
