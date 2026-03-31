FROM golang:1.25 AS builder

# Build context is the REPO ROOT: docker build -f auth_service/Dockerfile .
WORKDIR /app

# Copy go.work first
COPY go.work go.work.sum* ./

# Copy ALL go.mod files so go.work can resolve every module listed in 'use'
COPY api_gateway/go.mod ./api_gateway/
COPY auth_service/go.mod auth_service/go.sum ./auth_service/
COPY email_service/go.mod ./email_service/
COPY order_service/go.mod ./order_service/
COPY product_service/go.mod product_service/go.sum ./product_service/
COPY user_client/go.mod ./user_client/
COPY user_service/go.mod user_service/go.sum ./user_service/

# Download dependencies
WORKDIR /app/auth_service
RUN go mod download

# Copy source of modules actually needed at compile time
COPY auth_service/ /app/auth_service/
COPY user_service/ /app/user_service/

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth_service .

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/auth_service/auth_service /

CMD ["/auth_service"]