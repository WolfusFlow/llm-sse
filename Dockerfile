# syntax=docker/dockerfile:1.4

# 1. Build stage
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /llmsse ./cmd/server

# 2. Final stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /llmsse /llmsse

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/llmsse"]
