# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

# Runtime stage (minimal)
FROM scratch
COPY --from=builder /app/server /app/server
COPY stories /app/stories
COPY templates /app/templates
COPY static /app/static
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
