module github.com/tird4d/go-microservices/auth_service

go 1.24.2

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.7.3
	github.com/tird4d/go-microservices/user_service v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.17.3
	golang.org/x/crypto v0.32.0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
)

replace github.com/tird4d/go-microservices/user_service => ../user_service
