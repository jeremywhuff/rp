package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/jeremywhuff/rp"
	"github.com/jeremywhuff/rp/rpout"
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

// Purchase handler without rp

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

	// Create new order document

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

	// Send "order in progress" email to customer

	if err := emailClient.SendOrderInProgressAlert(body.CustomerID, body.SKU, body.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response

	c.JSON(http.StatusOK, gin.H{"message": "Purchase successful"})
}

// rp execution chains

func parseChain() *Chain {
	return First(
		// TODO: Should Bind take a pointer or a value?
		Bind(&PurchaseRequestBody{})).Then(
		CtxSet("req.body"))
}

func fetchCustomerChain() *Chain {
	return First(

		S(`fetch_customer_query(["req.body"].CustomerID) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				customerID := c.MustGet("req.body").(*PurchaseRequestBody).CustomerID
				query := map[string]any{
					"customer_id": customerID}
				return query, nil

			})).Then(

		rpout.MongoFindOne("mongo.client.database", "customers", rpout.MongoFindOneOptions{
			Result: CustomerDocument{}})).Then(

		CtxSet("mongo.document.customer"))
}

func fetchInventoryChain() *Chain {
	return First(

		S(`fetch_inventory_query(["req.body"].SKU) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				sku := c.MustGet("req.body").(*PurchaseRequestBody).SKU
				query := map[string]any{
					"sku": sku}
				return query, nil

			})).Then(

		rpout.MongoFindOne("mongo.client.database", "inventory", rpout.MongoFindOneOptions{
			Result: InventoryDocument{}})).Then(

		CtxSet("mongo.document.inventory"))
}

var ErrNotEnoughStock = errors.New("not enough stock")

func checkInventoryStockChain() *Chain {
	return MakeChain(

		S(`check_inventory_stock(["mongo.document.inventory"].Stock, ["req.body"].Quantity)`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				stock := c.MustGet("mongo.document.inventory").(*InventoryDocument).Stock
				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity

				if stock < quantity {
					return nil, ErrNotEnoughStock
				}
				return nil, nil

			})).Catch(BR, "Not enough stock")
}

func calculateTotalChain() *Chain {
	return MakeChain(

		S(`calculate_total(["req.body"].Quantity, ["mongo.document.inventory"].Price) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity
				price := c.MustGet("mongo.document.inventory").(*InventoryDocument).Price

				return quantity * price, nil

			})).Then(

		CtxSet("total"))
}

func runPaymentChain() *Chain {
	return MakeChain(

		S(`run_payment(["total"], ["mongo.document.customer"].WalletID)`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				paymentClient := c.MustGet("payment.client").(*PaymentClient)

				total := c.MustGet("total").(int)
				walletID := c.MustGet("mongo.document.customer").(*CustomerDocument).WalletID

				_, err := paymentClient.RunTransaction(total, walletID)
				if err != nil {
					return nil, err
				}
				return nil, nil
			}))
}
