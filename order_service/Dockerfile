# Build context is the REPO ROOT: docker build -f order_service/Dockerfile .
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go.work first
COPY go.work go.work.sum* ./

# Copy ALL go.mod files so go.work can resolve every module
COPY api_gateway/go.mod ./api_gateway/
COPY auth_service/go.mod ./auth_service/
COPY email_service/go.mod ./email_service/
COPY order_service/go.mod order_service/go.sum ./order_service/
COPY product_service/go.mod product_service/go.sum ./product_service/
COPY user_client/go.mod ./user_client/
COPY user_service/go.mod ./user_service/

WORKDIR /app/order_service
RUN go mod download

# Copy source of modules needed at compile time
COPY order_service/ /app/order_service/
COPY product_service/ /app/product_service/

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order_service .

# Stage 2: Runtime
FROM gcr.io/distroless/static

COPY --from=builder /app/order_service/order_service /order_service

EXPOSE 8082
CMD ["/order_service"]
