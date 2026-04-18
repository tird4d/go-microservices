# Build context is the REPO ROOT: docker build -f email_service/Dockerfile .
FROM golang:1.25

WORKDIR /app

# Copy go.work so the workspace module resolution works inside Docker
COPY go.work go.work.sum* ./

# Copy ALL go.mod files so go.work can resolve every module listed in 'use'
COPY api_gateway/go.mod ./api_gateway/
COPY auth_service/go.mod ./auth_service/
COPY email_service/go.mod email_service/go.sum ./email_service/
COPY order_service/go.mod ./order_service/
COPY product_service/go.mod ./product_service/
COPY user_client/go.mod ./user_client/
COPY user_service/go.mod ./user_service/

RUN cd email_service && go mod download

COPY email_service/ /app/email_service/

WORKDIR /app/email_service
RUN go build -o main .

CMD ["./main"]