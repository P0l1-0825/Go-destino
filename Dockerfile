FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /godestino ./cmd/api

# Runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /godestino .
COPY migrations/ ./migrations/

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./godestino"]
