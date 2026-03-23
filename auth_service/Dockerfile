FROM golang:1.24 AS builder

# Build context is the REPO ROOT (docker build -f auth_service/Dockerfile .)
# This allows the replace directive (../user_service) to resolve correctly
WORKDIR /app

# Copy both module definitions first (better layer caching)
COPY auth_service/go.mod auth_service/go.sum ./auth_service/
COPY user_service/go.mod user_service/go.sum ./user_service/

# Download dependencies (replace directive resolves because ../user_service exists)
WORKDIR /app/auth_service
RUN go mod download

# Copy full source of both modules
COPY auth_service/ /app/auth_service/
COPY user_service/ /app/user_service/

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth_service .

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/auth_service/auth_service /

CMD ["/auth_service"]