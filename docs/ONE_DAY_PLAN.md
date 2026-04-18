# One-Day Feature Sprint — April 18, 2026

**Goal:** Transform isolated services into a working e-commerce flow:
> User browses products → places an order → receives confirmation email

---

## Current State (What Already Exists)

| Component | Status |
|-----------|--------|
| RabbitMQ | ✅ In docker-compose, running |
| `user.registered` event | ✅ user_service publishes → email_service consumes (welcome email) |
| order_service listens to `user_exchange` | ✅ Saves user to in-memory store |
| product_service gRPC | ✅ Full CRUD: `GetProduct`, `ListProducts`, price + stock in model |
| email_service RabbitMQ consumer | ✅ Connected, needs `order.placed` consumer added |
| order_service HTTP API | ❌ Does not exist |
| order_service MongoDB storage | ❌ Only in-memory user store |
| order_service → product_service gRPC | ❌ No gRPC client |
| `order.placed` event | ❌ Not published |
| api_gateway → order_service routing | ❌ No routes |
| Frontend store page | ❌ Does not exist |

---

## Step-by-Step Plan

### Step 1 — order_service: Order Model + MongoDB Storage
**~45 min**

Files to create:
- `order_service/models/order.go` — Order struct (id, user_id, items[], total_price, status, created_at)
- `order_service/storage/mongo.go` — MongoDB connect + SaveOrder / GetOrdersByUser
- Update `order_service/go.mod` — add `go.mongodb.org/mongo-driver`

```
Order {
  ID         string
  UserID     string
  Items      []OrderItem { ProductID, Name, Price, Quantity }
  TotalPrice float64
  Status     string  // "pending" | "confirmed" | "cancelled"
  CreatedAt  time.Time
}
```

---

### Step 2 — order_service: gRPC Client → product_service
**~45 min**

Files to create:
- `order_service/proto/` — copy product.proto + generated pb.go files (or vendor them)
- `order_service/clients/product_client.go` — dial product-service:50053, call `GetProduct`

Used in: order creation — fetch product name + price + verify stock > 0

---

### Step 3 — order_service: HTTP API
**~1h**

Files to create:
- `order_service/handlers/order_handler.go`
  - `POST /orders` — create order (requires auth header with user_id)
  - `GET  /orders` — list orders for authenticated user
- `order_service/main.go` — refactor to start HTTP server (Gin) + keep RabbitMQ consumer

Request body for `POST /orders`:
```json
{
  "items": [
    { "product_id": "...", "quantity": 2 }
  ]
}
```

Response:
```json
{
  "order_id": "...",
  "total_price": 49.98,
  "status": "confirmed"
}
```

---

### Step 4 — order_service: Publish `order.placed` Event
**~30 min**

- After saving order to MongoDB, publish to `order_exchange` (fanout)
- Event payload:
```json
{
  "order_id": "...",
  "user_id": "...",
  "user_email": "...",
  "items": [...],
  "total_price": 49.98
}
```
- File: `order_service/events/publisher.go`

---

### Step 5 — email_service: Consume `order.placed` Event
**~30 min**

- Bind to `order_exchange`
- On message: log "Sending order confirmation to {email}" (real SMTP optional)
- File: add `order_consumer` function to `email_service/main.go`

> Real SMTP (Gmail/SendGrid) is optional — log output is enough for demo

---

### Step 6 — api_gateway: Route /api/v1/orders
**~30 min**

- Add `order-service` gRPC/HTTP client in api_gateway
- New routes:
  - `POST /api/v1/orders` → order_service (auth required)
  - `GET  /api/v1/orders` → order_service (auth required)
- File: `api_gateway/handlers/order_handler.go`

---

### Step 7 — docker-compose: Add order_service
**~15 min**

```yaml
order-service:
  build: ./order_service
  ports:
    - "8082:8082"
  env_file: ./order_service/.env
  depends_on:
    - mongo
    - rabbitmq
    - product-service
```

---

### Step 8 — Frontend: Product Store Page
**~2h**

Simple React pages in `frontend_service/src/`:
- `/products` — grid of product cards (name, price, image, "Add to Cart" button)
- `/cart` — list of items + "Place Order" button
- `/orders` — order history

API calls go through `api_gateway` at `/api/v1/products` and `/api/v1/orders`.

---

## End-to-End Flow When Done

```
Browser → POST /api/v1/orders
  → api_gateway (auth middleware validates JWT)
    → order_service HTTP handler
      → product_service gRPC: GetProduct (price + stock check)
        → save Order to MongoDB
          → publish "order.placed" to RabbitMQ order_exchange
            → email_service consumes → sends confirmation email
  ← 200 { order_id, total_price, status: "confirmed" }
```

---

## Order of Execution Today

- [x] Step 1 — Order model + MongoDB
- [x] Step 2 — Product gRPC client
- [x] Step 3 — HTTP API (POST/GET /orders)
- [x] Step 4 — Publish order.placed event
- [x] Step 5 — email_service consumes order.placed
- [x] Step 6 — api_gateway routes
- [x] Step 7 — docker-compose wiring
- [x] Step 8 — Frontend store UI

---

## Definition of Done

- `docker compose up` starts everything with no errors
- Can register → log in → list products → place order
- order_service logs show order saved to MongoDB
- email_service logs show "Sending order confirmation to {email}"
- Frontend shows product grid and order history
