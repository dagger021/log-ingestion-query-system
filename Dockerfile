FROM golang:1.25 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Copy source code
COPY . .

# Build binary
# RUN --mount=type=cache,target=/root/.cache/go-build \
#   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#   go build -o server ./cmd/server
RUN --mount=type=cache,target=/go/pkg/mod,id=gomodcache \
  --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server .
COPY --from=builder /app/schema.sql .
COPY --from=builder /app/templates ./templates

EXPOSE 3000

CMD ["./server"]
