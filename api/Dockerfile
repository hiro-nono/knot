# ---- Build stage ----
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

# ---- Run stage ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /app/bin/server .

USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
