FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod ./
# RUN go mod download

COPY go.mod ./
COPY go.sum ./
COPY vendor/ ./vendor/
COPY . .

RUN go build -mod=vendor -o main .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o user-service .

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static

COPY --from=builder /app/user-service /

CMD ["/user-service"]