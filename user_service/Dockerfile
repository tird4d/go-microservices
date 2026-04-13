# Stage 1: Build
# Build context is the REPO ROOT: docker build -f user_service/Dockerfile .
FROM golang:1.25 AS builder
WORKDIR /app

# Copy go.work first
COPY go.work go.work.sum* ./

# Copy ALL go.mod files so go.work can resolve every module listed in 'use'
COPY api_gateway/go.mod ./api_gateway/
COPY auth_service/go.mod ./auth_service/
COPY email_service/go.mod ./email_service/
COPY order_service/go.mod ./order_service/
COPY product_service/go.mod ./product_service/
COPY user_client/go.mod ./user_client/
COPY user_service/go.mod user_service/go.sum ./user_service/

WORKDIR /app/user_service
RUN go mod download

# Copy source of modules actually needed at compile time
COPY user_service/ /app/user_service/

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .



# Download grpc-health-probe binary
ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.48/grpc_health_probe-linux-amd64 /app/grpc-health-probe
RUN chmod +x /app/grpc-health-probe

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/user_service/user-service /
COPY --from=builder /app/grpc-health-probe /bin/grpc-health-probe

CMD ["/user-service"]