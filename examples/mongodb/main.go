package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
// - Run payment
// - Create shipment
// - Create order
// - Send email for receipt

// Database Documents

type CustomerDocument struct {
	ID         primitive.ObjectID `bson:"_id"`
	CustomerID string             `bson:"customer_id"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Email      string             `bson:"email"`
	WalletID   string             `bson:"wallet_id"`
}

type InventoryDocument struct {
	ID    primitive.ObjectID `bson:"_id"`
	SKU   string             `bson:"sku"`
	Name  string             `bson:"name"`
	Price int                `bson:"price"`
	Stock int                `bson:"stock"`
}

type OrderDocument struct {
	ID                  primitive.ObjectID `bson:"_id"`
	OrderDocumentFields `bson:",inline"`
}

type OrderDocumentFields struct {
	Customer primitive.ObjectID `bson:"customer"`
	Item     primitive.ObjectID `bson:"item"`
	Quantity int                `bson:"quantity"`
	Total    int                `bson:"total"`
}

// Dummy Clients

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

// Purchase Route

var Method = http.MethodPost
var Path = "/purchase"

type PurchaseRequestBody struct {
	CustomerID string `json:"customer_id"`
	SKU        string `json:"sku"`
	Quantity   int    `json:"quantity"`
}

func main() {

	// Connect to MongoDB

	mongoDBURI := "mongodb://localhost:27017"

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoDBURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// t := time.Now()

	// _, err = client.Database("rp_test").Collection("customers").InsertOne(context.Background(), map[string]any{"first_name": "Ted", "last_name": "Hooper"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println(time.Since(t))
}

func PurchaseHandler(mongoClient *mongo.Client, paymentClient *PaymentClient, shippingClient *ShippingClient, emailClient *EmailClient, c *gin.Context) {

	// Parse request body

	var body PurchaseRequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch customer

	customer := CustomerDocument{}
	err = mongoClient.Database("rp_test").Collection("customers").FindOne(context.Background(),
		map[string]any{
			"customer_id": body.CustomerID,
		}).Decode(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch inventory

	item := InventoryDocument{}
	err = mongoClient.Database("rp_test").Collection("inventory").FindOne(context.Background(),
		map[string]any{
			"sku": body.SKU,
		}).Decode(&item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check inventory stock

	if item.Stock < body.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock"})
		return
	}

	// Run payment

	total := body.Quantity * item.Price

	_, err = paymentClient.RunTransaction(total, customer.WalletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create shipment

	_, err = shippingClient.CreateShipment(body.CustomerID, body.SKU, body.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create order

	order := OrderDocumentFields{
		Customer: customer.ID,
		Item:     item.ID,
		Quantity: body.Quantity,
		Total:    total,
	}

	_, err = mongoClient.Database("rp_test").Collection("orders").InsertOne(context.Background(), order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send email for receipt

	if err := emailClient.SendOrderInProgressAlert(body.CustomerID, body.SKU, body.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response

	c.JSON(http.StatusOK, gin.H{"message": "Order created"})
}
