# Product Service Handlers - Testing Guide

## ✅ Implemented Handlers (with Mock Data)

All handlers are implemented with in-memory mock data storage.

### **1. CreateProduct** 
Creates a new product

**Example:**
```json
{
  "name": "New Laptop",
  "description": "Gaming laptop",
  "price": 1499.99,
  "category": "electronics",
  "stock": 10,
  "image_url": "https://example.com/laptop.jpg"
}
```

### **2. GetProduct**
Get a single product by ID

**Pre-loaded Mock IDs:**
- `"1"` - Laptop Pro 15 ($1299.99)
- `"2"` - Wireless Mouse ($29.99)
- `"3"` - Programming Book ($45.00)

### **3. UpdateProduct**
Update an existing product

### **4. DeleteProduct**
Delete a product by ID

### **5. ListProducts**
List all products with pagination
- Default page: 1
- Default pageSize: 10

### **6. GetProductsByCategory**
Filter products by category with pagination

**Available Categories in Mock Data:**
- `electronics` (2 products)
- `books` (1 product)

---

## 🧪 Test Using Go Client

Create a test file:

```bash
cd /home/tirdad/Projects/go-microservices/product_service
touch test_client.go
```

**test_client.go:**
```go
package main

import (
    "context"
    "log"
    "time"

    pb "github.com/tird4d/go-microservices/product_service/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to product service
    conn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewProductServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()

    // Test 1: Get product
    log.Println("=== Test 1: Get Product ===")
    product, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: "1"})
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Product: %s - $%.2f", product.Name, product.Price)
    }

    // Test 2: Create product
    log.Println("\n=== Test 2: Create Product ===")
    newProduct, err := client.CreateProduct(ctx, &pb.CreateProductRequest{
        Name:        "Mechanical Keyboard",
        Description: "RGB mechanical gaming keyboard",
        Price:       149.99,
        Category:    "electronics",
        Stock:       50,
        ImageUrl:    "https://example.com/keyboard.jpg",
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Created: %s (ID: %s)", newProduct.Name, newProduct.Id)
    }

    // Test 3: List all products
    log.Println("\n=== Test 3: List Products ===")
    listResp, err := client.ListProducts(ctx, &pb.ListProductsRequest{
        Page:     1,
        PageSize: 10,
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Total products: %d", listResp.Total)
        for _, p := range listResp.Products {
            log.Printf("  - %s: $%.2f", p.Name, p.Price)
        }
    }

    // Test 4: Get by category
    log.Println("\n=== Test 4: Get Products by Category ===")
    categoryResp, err := client.GetProductsByCategory(ctx, &pb.GetProductsByCategoryRequest{
        Category: "electronics",
        Page:     1,
        PageSize: 10,
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Electronics category: %d products", categoryResp.Total)
        for _, p := range categoryResp.Products {
            log.Printf("  - %s", p.Name)
        }
    }

    // Test 5: Update product
    log.Println("\n=== Test 5: Update Product ===")
    updatedProduct, err := client.UpdateProduct(ctx, &pb.UpdateProductRequest{
        Id:          "1",
        Name:        "Laptop Pro 15 - Updated",
        Description: "Updated description",
        Price:       1199.99,
        Category:    "electronics",
        Stock:       30,
        ImageUrl:    "https://example.com/laptop-updated.jpg",
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Updated: %s - $%.2f", updatedProduct.Name, updatedProduct.Price)
    }

    // Test 6: Delete product
    log.Println("\n=== Test 6: Delete Product ===")
    deleteResp, err := client.DeleteProduct(ctx, &pb.DeleteProductRequest{Id: "2"})
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Delete success: %v - %s", deleteResp.Success, deleteResp.Message)
    }

    log.Println("\n✅ All tests completed!")
}
```

**Run test:**
```bash
go run test_client.go
```

---

## 📊 Mock Data Included

**3 pre-loaded products:**

1. **Laptop Pro 15** (ID: "1")
   - Price: $1299.99
   - Category: electronics
   - Stock: 25

2. **Wireless Mouse** (ID: "2")
   - Price: $29.99
   - Category: electronics
   - Stock: 150

3. **Programming Book** (ID: "3")
   - Price: $45.00
   - Category: books
   - Stock: 80

---

## 🔄 Next Steps

1. ✅ **Test with client** - Run test_client.go
2. ✅ **Add to API Gateway** - Create HTTP→gRPC bridge
3. ✅ **Replace mock with MongoDB** - Implement repository layer
4. ✅ **Add to Docker Compose** - Containerize
5. ✅ **Deploy to Kubernetes** - Create Helm chart

---

## 💡 Features Implemented

✅ Full CRUD operations
✅ Pagination support
✅ Category filtering
✅ Structured logging with zap
✅ Timestamp tracking (createdAt, updatedAt)
✅ Error handling
✅ In-memory data storage (easy to test)

