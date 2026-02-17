# Build stage
FROM golang:1.24.2 AS builder

WORKDIR /app

# RUN go mod download

COPY go.mod ./
COPY go.sum ./
COPY vendor/ ./vendor/

# Copy the rest of the application code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o main .

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Use a non-root user for better security
USER nonroot:nonroot

CMD ["./main"]
