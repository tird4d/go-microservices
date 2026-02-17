# Build stage
FROM golang:1.25 AS builder

WORKDIR /workspace

# Copy all services (needed for replace directives in go.mod)
COPY . .

# Build from api_gateway directory
WORKDIR /workspace/api_gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /workspace/api_gateway/main .

# Use a non-root user for better security
USER nonroot:nonroot

CMD ["./main"]
