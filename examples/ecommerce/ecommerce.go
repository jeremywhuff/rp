package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/jeremywhuff/rp"
	"github.com/jeremywhuff/rp/modules/rpmongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This example demonstrates a simplified ecommerce endpoint to purchase an item.
// See tutorial for more details. https://medium.com/@jeremywhuff/950a10c3c31f
//
// The endpoint will be implemented in four different ways:
// A) Without rp,
// B) As a direct migration into rp,
// C) A tidied up implementation in rp, and
// D) With concurrency optimizations in rp.
//
// Every implementation will take these steps:
// 1) Fetch customer
// 2) Fetch inventory
// 3) Check inventory stock
// 4) Run payment
// 5) Create shipment
// 6) Create order
// 7) Send email for receipt

// Purchase Route - These variables define the example route

var Method = http.MethodPost
var Path = "/purchase"

type PurchaseRequestBody struct {
	CustomerID string `json:"customer_id"`
	SKU        string `json:"sku"`
	Quantity   int    `json:"quantity"`
}

func main() {

	// Connect to MongoDB

	// Set up a DB to connect to, for instance a local one like this - https://brandonblankenstein.medium.com/install-and-run-mongodb-on-mac-1604ae750e57
	// MAKE SURE this DB is empty, as it will be cleared and populated with dummy data
	mongoDBURI := "" //"mongodb://localhost:27017"

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoDBURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Set up database

	err = setUpDB(mongoClient.Database("rp_test"))
	if err != nil {
		log.Fatal(err)
	}

	// Set up dummy clients

	paymentClient := &PaymentClient{}
	shippingClient := &ShippingClient{}
	emailClient := &EmailClient{}

	// Set up and run router (Comment and uncomment as necessary)
	r := gin.Default()

	// A) Without rp
	r.POST(Path, PurchaseHandler(mongoClient, paymentClient, shippingClient, emailClient))

	// B) Direct migration to rp
	// r.POST(Path, PurchaseHandlerDirectMigrationToRP(mongoClient, paymentClient, shippingClient, emailClient))

	// C) Tidied up implementation in rp
	// r.POST(Path, MiddlewareForRPHandlers(mongoClient, paymentClient, shippingClient, emailClient), PurchaseHandlerWithRP(mongoClient, paymentClient, shippingClient, emailClient, false))

	// D) With concurrency optimizations in rp
	// r.POST(Path, MiddlewareForRPHandlers(mongoClient, paymentClient, shippingClient, emailClient), PurchaseHandlerWithRP(mongoClient, paymentClient, shippingClient, emailClient, true))

	r.Run(":8081")
}

// *** MongoDB ***

// Collections are "customers", "inventory", and "orders".

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
	Price int                `bson:"price"` // in cents
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
	Total    int                `bson:"total"` // in cents
}

// Database initialization

func setUpDB(db *mongo.Database) error {

	// Clear all customers, add a bunch of dummy records, and create the test customer
	db.Collection("customers").DeleteMany(context.Background(), map[string]any{})
	dummyCustomers := make([]any, 100000)
	for i := range dummyCustomers {
		indexString := fmt.Sprintf("%06d", i)
		dummyCustomers[i] = CustomerDocument{
			ID:         primitive.NewObjectID(),
			CustomerID: "C" + indexString,
			FirstName:  "First" + indexString,
			LastName:   "Last" + indexString,
			Email:      "first" + indexString + "@example.com",
			WalletID:   "W" + indexString,
		}
	}
	db.Collection("customers").InsertMany(context.Background(), dummyCustomers)
	db.Collection("customers").InsertOne(context.Background(), CustomerDocument{
		ID:         primitive.NewObjectID(),
		CustomerID: "C975310",
		FirstName:  "Sandra",
		LastName:   "Hernandez",
		Email:      "sandra.hernandez@example.com",
		WalletID:   "W246802",
	})

	// Clear all inventory items, add a bunch of dummy records, and create the test item
	db.Collection("inventory").DeleteMany(context.Background(), map[string]any{})
	dummyItems := make([]any, 50000)
	for i := range dummyItems {
		indexString := fmt.Sprintf("%06d", i)
		dummyItems[i] = InventoryDocument{
			ID:    primitive.NewObjectID(),
			SKU:   "SKU" + indexString,
			Name:  "Product" + indexString,
			Price: 2000,
			Stock: 15,
		}
	}
	db.Collection("inventory").InsertMany(context.Background(), dummyItems)
	_, err := db.Collection("inventory").InsertOne(context.Background(), InventoryDocument{
		ID:    primitive.NewObjectID(),
		SKU:   "SKU159260",
		Name:  "Wonder Widget",
		Price: 3500,
		Stock: 5,
	})
	if err != nil {
		return err
	}

	// Clear all orders
	db.Collection("orders").DeleteMany(context.Background(), map[string]any{})
	dummyOrders := make([]any, 200000)
	for i := range dummyItems {
		dummyOrders[i] = OrderDocument{
			ID: primitive.NewObjectID(),
		}
	}
	db.Collection("orders").InsertMany(context.Background(), dummyOrders)

	return nil
}

// *** API Clients ***

// Dummy clients for external APIs: payment, shipping, and email.
// They always succeed after a delay.

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

// *** Handlers ***

// Purchase handler without rp
func PurchaseHandler(mongoClient *mongo.Client, paymentClient *PaymentClient, shippingClient *ShippingClient, emailClient *EmailClient) gin.HandlerFunc {

	return func(c *gin.Context) {

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
}

// Required middleware for all rp handlers
func MiddlewareForRPHandlers(mongoClient *mongo.Client, paymentClient *PaymentClient, shippingClient *ShippingClient, emailClient *EmailClient) gin.HandlerFunc {

	return func(c *gin.Context) {

		c.Set("mongo.client.database", mongoClient.Database("rp_test"))
		c.Set("payment.client", paymentClient)
		c.Set("shipping.client", shippingClient)
		c.Set("email.client", emailClient)

		c.Next()
	}
}

// Purchase handler with rp, direct migration
func PurchaseHandlerDirectMigrationToRP(mongoClient *mongo.Client, paymentClient *PaymentClient, shippingClient *ShippingClient, emailClient *EmailClient) gin.HandlerFunc {

	parse := MakeChain(S(
		FuncStr("parse")+CtxOutStr("req.body"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			var body PurchaseRequestBody
			err := c.ShouldBindJSON(&body)
			if err != nil {
				return nil, err
			}

			c.Set("req.body", &body)
			return nil, nil
		}))

	fetchCustomer := MakeChain(S(
		FuncStr("fetch_customer", "req.body")+CtxOutStr("mongo.document.customer"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			body := c.MustGet("req.body").(*PurchaseRequestBody)

			customer := CustomerDocument{}
			err := mongoClient.Database("rp_test").Collection("customers").FindOne(context.Background(),
				map[string]any{
					"customer_id": body.CustomerID,
				}).Decode(&customer)
			if err != nil {
				return nil, err
			}

			c.Set("mongo.document.customer", &customer)
			return nil, nil
		}))

	fetchInventory := MakeChain(S(
		FuncStr("fetch_inventory", "req.body")+CtxOutStr("mongo.document.inventory"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			body := c.MustGet("req.body").(*PurchaseRequestBody)

			item := InventoryDocument{}
			err := mongoClient.Database("rp_test").Collection("inventory").FindOne(context.Background(),
				map[string]any{
					"sku": body.SKU,
				}).Decode(&item)
			if err != nil {
				return nil, err
			}

			c.Set("mongo.document.inventory", &item)
			return nil, nil
		}))

	checkStock := MakeChain(S(
		FuncStr("check_stock", "mongo.document.inventory", "req.body"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			item := c.MustGet("mongo.document.inventory").(*InventoryDocument)
			body := c.MustGet("req.body").(*PurchaseRequestBody)

			if item.Stock < body.Quantity {
				return nil, errors.New("Not enough stock")
			}

			return nil, nil
		}))

	runPayment := MakeChain(S(
		FuncStr("run_payment", "req.body", "mongo.document.inventory", "mongo.document.customer")+CtxOutStr("total"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			body := c.MustGet("req.body").(*PurchaseRequestBody)
			item := c.MustGet("mongo.document.inventory").(*InventoryDocument)
			customer := c.MustGet("mongo.document.customer").(*CustomerDocument)

			total := body.Quantity * item.Price
			c.Set("total", total)

			_, err := paymentClient.RunTransaction(total, customer.WalletID)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}))

	createShipment := MakeChain(S(
		FuncStr("create_shipment", "req.body"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			body := c.MustGet("req.body").(*PurchaseRequestBody)

			_, err := shippingClient.CreateShipment(body.CustomerID, body.SKU, body.Quantity)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}))

	createOrder := MakeChain(S(
		FuncStr("create_order", "mongo.document.customer", "mongo.document.inventory", "req.body", "total"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			customer := c.MustGet("mongo.document.customer").(*CustomerDocument)
			item := c.MustGet("mongo.document.inventory").(*InventoryDocument)
			body := c.MustGet("req.body").(*PurchaseRequestBody)
			total := c.MustGet("total").(int)

			order := OrderDocumentFields{
				Customer: customer.ID,
				Item:     item.ID,
				Quantity: body.Quantity,
				Total:    total,
			}

			_, err := mongoClient.Database("rp_test").Collection("orders").InsertOne(context.Background(), order)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}))

	sendEmail := MakeChain(S(
		FuncStr("send_email", "req.body"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			body := c.MustGet("req.body").(*PurchaseRequestBody)

			if err := emailClient.SendOrderInProgressAlert(body.CustomerID, body.SKU, body.Quantity); err != nil {
				return nil, err
			}

			return nil, nil
		}))

	respond := MakeChain(S(
		FuncStr("respond"),
		func(in any, c *gin.Context, lgr Logger) (any, error) {
			res := Response{
				Code: http.StatusOK,
				Obj:  gin.H{"message": "Purchase successful"},
			}
			return &res, nil
		}))

	// Build the full pipeline chain
	pipeline := InSequence(
		parse,
		fetchCustomer,
		fetchInventory,
		checkStock,
		runPayment,
		createShipment,
		createOrder,
		sendEmail,
		respond,
	)

	return MakeGinHandlerFunc(pipeline, DefaultLogger{})
}

// Tidied up handler in rp, which can be set to run with or without concurrency optimizations
func PurchaseHandlerWithRP(withConcurrency bool) gin.HandlerFunc {

	// First: Parse request body
	parse := First(
		Bind(&PurchaseRequestBody{})).Then(
		CtxSet("req.body"))

	// 1) Fetch customer
	fetchCustomer := First(

		S(`fetch_customer_query(["req.body"]) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				customerID := c.MustGet("req.body").(*PurchaseRequestBody).CustomerID
				query := map[string]any{
					"customer_id": customerID}
				return query, nil

			})).Then(

		rpmongo.MongoFindOne("mongo.client.database", "customers", rpmongo.MongoFindOneOptions{
			Result: &CustomerDocument{}})).Then(

		CtxSet("mongo.document.customer"))

	// 2) Fetch inventory
	fetchInventory := First(

		S(`fetch_inventory_query(["req.body"]) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				sku := c.MustGet("req.body").(*PurchaseRequestBody).SKU
				query := map[string]any{
					"sku": sku}
				return query, nil

			})).Then(

		rpmongo.MongoFindOne("mongo.client.database", "inventory", rpmongo.MongoFindOneOptions{
			Result: &InventoryDocument{}})).Then(

		CtxSet("mongo.document.inventory"))

	// 3) Check inventory stock
	checkStock := MakeChain(

		S(`check_inventory_stock(["mongo.document.inventory"], ["req.body"])`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				stock := c.MustGet("mongo.document.inventory").(*InventoryDocument).Stock
				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity

				if stock < quantity {
					return nil, errors.New("not enough stock")
				}
				return nil, nil

			})).Catch(BR, "Not enough stock")

	// 4) Run payment

	// 4a) Calculate total
	calculateTotal := MakeChain(

		S(`calculate_total(["req.body"], ["mongo.document.inventory"]) =>`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity
				price := c.MustGet("mongo.document.inventory").(*InventoryDocument).Price

				return quantity * price, nil

			})).Then(

		CtxSet("total"))

	// 4b) Run payment with paymentClient
	runPayment := MakeChain(

		S(`run_payment(["total"], ["mongo.document.customer"])`,
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

	// 5) Create shipment
	createShipment := MakeChain(

		S(`create_shipment(["req.body"])`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				shippingClient := c.MustGet("shipping.client").(*ShippingClient)

				customerID := c.MustGet("req.body").(*PurchaseRequestBody).CustomerID
				sku := c.MustGet("req.body").(*PurchaseRequestBody).SKU
				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity

				_, err := shippingClient.CreateShipment(customerID, sku, quantity)
				if err != nil {
					return nil, err
				}
				return nil, nil
			}))

	// 6) Create order
	createOrder := First(

		S(`create_new_order(["mongo.document.customer"], ["mongo.document.inventory"], ["req.body"], ["total"])`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				customerID := c.MustGet("mongo.document.customer").(*CustomerDocument).ID
				itemID := c.MustGet("mongo.document.inventory").(*InventoryDocument).ID
				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity
				total := c.MustGet("total").(int)

				order := OrderDocumentFields{
					Customer: customerID,
					Item:     itemID,
					Quantity: quantity,
					Total:    total,
				}
				return order, nil
			})).Then(

		rpmongo.MongoInsert("mongo.client.database", "orders"))

	// 7) Send email for receipt
	sendOrderInProgressAlert := MakeChain(

		S(`send_order_in_progress_alert(["req.body"])`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				emailClient := c.MustGet("email.client").(*EmailClient)

				customerID := c.MustGet("req.body").(*PurchaseRequestBody).CustomerID
				sku := c.MustGet("req.body").(*PurchaseRequestBody).SKU
				quantity := c.MustGet("req.body").(*PurchaseRequestBody).Quantity

				err := emailClient.SendOrderInProgressAlert(customerID, sku, quantity)
				if err != nil {
					return nil, err
				}
				return nil, nil
			}))

	// Last: Return response
	successResponse := MakeChain(

		S(`success_response()`,
			func(in any, c *gin.Context, lgr Logger) (any, error) {

				res := Response{
					Code: http.StatusOK,
					Obj:  gin.H{"message": "Purchase successful"},
				}
				return &res, nil
			}))

	// Build the full pipeline chain
	pipeline := &Chain{}
	if !withConcurrency {
		pipeline = InSequence(
			parse,
			fetchCustomer,
			fetchInventory,
			checkStock,
			calculateTotal,
			runPayment,
			createShipment,
			createOrder,
			sendOrderInProgressAlert,
			successResponse,
		)
	} else {
		pipeline = InSequence(
			parse,
			InParallel(
				fetchCustomer,
				fetchInventory),
			checkStock,
			calculateTotal,
			runPayment,
			InParallel(
				createShipment,
				createOrder,
				sendOrderInProgressAlert),
			successResponse,
		)
	}

	return MakeGinHandlerFunc(pipeline, DefaultLogger{})
}
