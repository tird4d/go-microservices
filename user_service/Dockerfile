FROM golang:1.24

WORKDIR /app

ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 /bin/grpc-health-probe
RUN chmod +x /bin/grpc-health-probe

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]