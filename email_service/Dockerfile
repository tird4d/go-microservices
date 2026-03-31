# Build context is the REPO ROOT: docker build -f email_service/Dockerfile .
FROM golang:1.25

WORKDIR /app

# Copy go.work so the workspace module resolution works inside Docker
COPY go.work go.work.sum* ./

COPY email_service/go.mod ./email_service/
RUN cd email_service && go mod download

COPY email_service/ /app/email_service/

WORKDIR /app/email_service
RUN go build -o main .

CMD ["./main"]