package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Endpoint is to purchase an item.
// Collections are customers, inventory, and orders.
// Clients for external APIs are payment, shipping, and email.
//
// Steps:
// - Fetch customer
// - Fetch inventory
// - Check inventory stock
// - Check customer balance
// - Run payment to re-up credits, if needed
// - Create shipment
// - Create order
// - Send email for receipt

type CustomerDocument struct {
	ID         primitive.ObjectID `bson:"_id"`
	CustomerID string             `bson:"customer_id"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Credits    int                `bson:"credits"`
	Email      string             `bson:"email"`
	WalletID   string             `bson:"wallet_id"`
}

type InventoryDocument struct {
	ID    primitive.ObjectID `bson:"_id"`
	SKU   string             `bson:"sku"`
	Name  string             `bson:"name"`
	Stock int                `bson:"stock"`
}

type OrderDocument struct {
	ID       primitive.ObjectID `bson:"_id"`
	OrderID  string             `bson:"order_id"`
	Customer primitive.ObjectID `bson:"customer"`
	SKU      string             `bson:"inventory"`
	Quantity int                `bson:"quantity"`
	Total    int                `bson:"total"`
}

// Clients

type PaymentClient struct{}

// Charges the customer's card on file
func (cl *PaymentClient) RunTransaction(amount int, walletID string) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return "5678", nil
}

type ShippingClient struct{}

// Initializes a shipment with the shipping provider
func (cl *ShippingClient) CreateShipment(customerID string, SKU string, quantity int) (string, error) {
	time.Sleep(70 * time.Millisecond)
	return "1234", nil
}

type EmailClient struct{}

// Sends an email to the customer to say their order is in progress
func (cl *EmailClient) SendOrderInProgressAlert(customerID string, SKU string, quantity int) error {
	time.Sleep(80 * time.Millisecond)
	return nil
}

func main() {

	mongoDBURI := "mongodb://localhost:27017"

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoDBURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	t := time.Now()

	_, err = client.Database("rp_test").Collection("customers").InsertOne(context.Background(), map[string]any{"first_name": "Ted", "last_name": "Hooper"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(time.Since(t))
}
