# Stage 1: Build
FROM golang:1.24 AS builder
WORKDIR /app

# Copy dependencies and source
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .



# Download grpc-health-probe binary
ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 /app/grpc-health-probe
RUN chmod +x /app/grpc-health-probe

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/user-service /
COPY --from=builder /app/grpc-health-probe /bin/grpc-health-probe

CMD ["/user-service"]