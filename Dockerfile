# Build the Go binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# Minimal runtime image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/server ./server
EXPOSE 8080
CMD ["./server"]
