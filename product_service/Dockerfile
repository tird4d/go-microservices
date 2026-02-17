FROM golang:1.25 AS builder

WORKDIR /app

# Copy dependencies first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product_service .

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/product_service /

CMD ["/product_service"]