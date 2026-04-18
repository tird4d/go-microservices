package models

import "time"

type OrderItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Name      string  `bson:"name"       json:"name"`
	Price     float64 `bson:"price"      json:"price"`
	Quantity  int32   `bson:"quantity"   json:"quantity"`
}

type Order struct {
	ID         string      `bson:"_id"         json:"id"`
	UserID     string      `bson:"user_id"     json:"user_id"`
	UserEmail  string      `bson:"user_email"  json:"user_email"`
	Items      []OrderItem `bson:"items"       json:"items"`
	TotalPrice float64     `bson:"total_price" json:"total_price"`
	Status     string      `bson:"status"      json:"status"` // pending | confirmed | cancelled
	CreatedAt  time.Time   `bson:"created_at"  json:"created_at"`
}
