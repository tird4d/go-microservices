FROM golang:1.25 AS builder

WORKDIR /app

# Copy dependencies first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product_service .

# Download grpc-health-probe binary (same version as user-service)
ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 /app/grpc-health-probe
RUN chmod +x /app/grpc-health-probe

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/product_service /
COPY --from=builder /app/grpc-health-probe /bin/grpc-health-probe

CMD ["/product_service"]