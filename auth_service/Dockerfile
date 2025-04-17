FROM golang:1.24

WORKDIR /app

COPY go.mod ./
# RUN go mod download

COPY go.mod ./
COPY go.sum ./
COPY vendor/ ./vendor/
COPY . .

RUN go build -mod=vendor -o main .

CMD ["./main"]