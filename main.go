package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	client "github.com/promotedai/promoted-go-delivery-client/delivery"
	"github.com/promotedai/schema/generated/go/proto/common"
	"github.com/promotedai/schema/generated/go/proto/delivery"
	"google.golang.org/protobuf/types/known/structpb"
)

// Product is an example Proto struct.
type Product struct {
	ID    string
	Name  string
	Price int
}

type Config struct {
	MetricsApiEndpointUrl     string
	MetricsApiKey             string
	DeliveryApiEndpointUrl    string
	DeliveryApiKey            string
	OnlyLog                   bool
	ShadowTrafficDeliveryRate float64
	BlockingShadowTraffic     bool
}

func main() {
	// Parse environment variables
	config := Config{
		MetricsApiEndpointUrl:     os.Getenv("METRICS_API_ENDPOINT_URL"),
		MetricsApiKey:             os.Getenv("METRICS_API_KEY"),
		DeliveryApiEndpointUrl:    os.Getenv("DELIVERY_API_ENDPOINT_URL"),
		DeliveryApiKey:            os.Getenv("DELIVERY_API_KEY"),
		OnlyLog:                   parseBoolEnv("ONLY_LOG", false),
		ShadowTrafficDeliveryRate: parseFloatEnv("SHADOW_TRAFFIC_DELIVERY_RATE", 0.0),
		BlockingShadowTraffic:     parseBoolEnv("BLOCKING_SHADOW_TRAFFIC", false),
	}

	// Validate arguments
	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize the client (placeholder)
	client, err := NewPromotedDeliveryClient(config)
	if err != nil {
		fmt.Println("Error initializing PromotedDeliveryClient")
		panic(err)
	}

	// Retrieve products
	products := getProducts()

	// Create a map of products to reorder after Promoted ranking.
	productsMap := make(map[string]*Product, len(products))
	for _, product := range products {
		productsMap[product.ID] = product
	}

	// Build request
	requestInsertions := newTestRequestInsertions(products)
	req, err := newTestRequest(requestInsertions, config.OnlyLog)
	if err != nil {
		fmt.Println("newTestRequest failed")
		panic(err)
	}

	// Call the Promoted delivery API.
	response, err := client.Deliver(req)
	if err != nil {
		fmt.Println("Delivery called failed")
		panic(err)
	}

	// Apply Promoted's re-ranking to the products.
	fmt.Printf("Execution server: %s\n", response.ExecutionServer)
	fmt.Printf("Client request ID: %s\n", response.ClientRequestID)
	fmt.Printf("Response\n")
	for _, insertion := range response.Response.Insertion {
		rerankedProducts, ok := productsMap[insertion.ContentId]
		if ok {
			fmt.Printf("%v\n", rerankedProducts)
		} else {
			// Can happen if the Delivery API is configured to return cached items.
			fmt.Printf("%v\n", insertion.ContentId)
		}
	}
}

func validateConfig(config Config) error {
	if config.MetricsApiEndpointUrl == "" {
		return errors.New("metricsApiEndpointUrl needs to be specified")
	}
	if config.MetricsApiKey == "" {
		return errors.New("metricsApiKey needs to be specified")
	}
	if config.DeliveryApiEndpointUrl == "" {
		return errors.New("deliveryApiEndpointUrl needs to be specified")
	}
	if config.DeliveryApiKey == "" {
		return errors.New("deliveryApiKey needs to be specified")
	}
	return nil
}

func newTestRequest(insertions []*delivery.Insertion, onlyLog bool) (*client.DeliveryRequest, error) {
	// Request-level parameters go here.
	requestPropsValue, err := structpb.NewValue(map[string]any{
		"category": "topic",
		"priceMin": 10.0,
	})
	if err != nil {
		return nil, err
	}
	return &client.DeliveryRequest{
		Request: &delivery.Request{
			UserInfo: &common.UserInfo{
				AnonUserId: "testAnonUserId1",
				UserId:     "testUserId1",
			},
			UseCase:     delivery.UseCase_SEARCH,
			SearchQuery: "query",
			Paging: &delivery.Paging{
				Starting: &delivery.Paging_Offset{Offset: 0},
				Size:     3,
			},
			DisablePersonalization: false,
			Properties: &common.Properties{
				StructField: &common.Properties_Struct{
					Struct: requestPropsValue.GetStructValue(),
				},
			},
			Insertion: insertions,
		},
		OnlyLog: onlyLog,
	}, nil
}

func newTestRequestInsertions(products []*Product) []*delivery.Insertion {
	// Build request
	// Create insertions for each product for the Promtoed delivery request.
	insertions := make([]*delivery.Insertion, 0, len(products))
	for _, product := range products {
		// This example sends price as a dynamic property.  If price is static, it can be sent through Content Store.
		insertionPropsValue, err := structpb.NewValue(map[string]any{
			"price": product.Price,
		})
		if err != nil {
			fmt.Println("insertion structpb.NewValue failed")
			panic(err)
		}
		insertions = append(insertions, &delivery.Insertion{
			ContentId: product.ID,
			Properties: &common.Properties{
				StructField: &common.Properties_Struct{
					Struct: insertionPropsValue.GetStructValue(),
				},
			},
		})
	}
	return insertions
}

func parseBoolEnv(key string, defaultValue bool) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseFloatEnv(key string, defaultValue float64) float64 {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func NewPromotedDeliveryClient(config Config) (*client.PromotedDeliveryClient, error) {
	return client.NewPromotedDeliveryClientBuilder().
		WithDeliveryEndpoint(config.DeliveryApiEndpointUrl).
		WithDeliveryAPIKey(config.DeliveryApiKey).
		WithDeliveryTimeoutMillis(1000).
		WithMetricsEndpoint(config.MetricsApiEndpointUrl).
		WithMetricsAPIKey(config.MetricsApiKey).
		WithMetricsTimeoutMillis(1000).
		WithAcceptsGzip(true).
		WithAPIFactory(&client.DefaultAPIFactory{}).
		Build()
}

func getProducts() []*Product {
	return []*Product{
		{
			ID:    "1",
			Name:  "Product 1",
			Price: 100,
		},
		{
			ID:    "2",
			Name:  "Product 2",
			Price: 200,
		},
	}
}
