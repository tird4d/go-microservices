FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
# RUN go mod download

# Copy dependencies and source
COPY go.mod go.sum ./
# RUN go mod download
RUN go mod vendor
COPY vendor/ ./vendor/
COPY . .

RUN go build -mod=vendor -o main .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o auth_service .

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/auth_service /

CMD ["/auth_service"]