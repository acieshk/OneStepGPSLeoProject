package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config represents the configuration structure for the application
type Config struct {
	ServerPort     string `json:"server_port"`
	MongoDBURL     string `json:"mongodb_url"`
	MongoDBPort    string `json:"mongodb_port"`
	DatabaseName   string `json:"database_name"`
	CollectionName string `json:"collection_name"`
	APIKey         string `json:"api_key"` // API key for external API access
	APIURL         string `json:"api_url"` // Base URL for the API
	FrontendURL    string `json:"frontend_url"`
	FrontendPort   string `json:"frontend_port"`
}

// Global variable to hold the MongoDB client
var client *mongo.Client

// loadConfig reads and parses the configuration file
func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	return config, err
}

// initMongoDB initializes the MongoDB client and establishes a connection
func initMongoDB(mongoURI string) error {
	var err error
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	fmt.Println("Connected to MongoDB!")
	return nil
}

func fetchAndStoreDevices(c *gin.Context, config Config) {
	fmt.Println("Starting fetchAndStoreDevices...")
	apiURL := fmt.Sprintf("%s%s", config.APIURL, config.APIKey)
	fmt.Printf("Fetching from API URL: %s\n", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("Error fetching from API: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from API"})
		return
	}
	defer resp.Body.Close()

	// Read the raw JSON response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response"})
		return
	}
	fmt.Printf("Raw API response received, length: %d bytes\n", len(body))

	// Parse JSON into a map
	var response struct {
		ResultList []map[string]interface{} `json:"result_list"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	// Get the devices collection
	collection := client.Database(config.DatabaseName).Collection(config.CollectionName)

	// Clear existing devices
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		fmt.Printf("Error clearing collection: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing devices"})
		return
	}
	fmt.Printf("Deleted %d existing documents\n", deleteResult.DeletedCount)

	// Store each device directly
	for i, device := range response.ResultList {
		_, err := collection.InsertOne(context.TODO(), device)
		if err != nil {
			fmt.Printf("Error inserting device %d: %v\n", i, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store device"})
			return
		}
	}

	fmt.Printf("Successfully stored %d devices\n", len(response.ResultList))
	c.JSON(http.StatusOK, gin.H{
		"message": "Devices stored successfully",
		"count":   len(response.ResultList),
	})
}

func getDevices(c *gin.Context, config Config) {
	fmt.Println("Getting devices from MongoDB...")
	collection := client.Database(config.DatabaseName).Collection(config.CollectionName)

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Printf("Error finding devices: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve devices"})
		return
	}
	defer cursor.Close(context.TODO())

	// Use a slice of raw documents instead of key-value pairs
	var devices []bson.M
	if err = cursor.All(context.TODO(), &devices); err != nil {
		fmt.Printf("Error decoding devices: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode devices"})
		return
	}

	fmt.Printf("Found %d devices\n", len(devices))
	c.JSON(http.StatusOK, gin.H{
		"result_list": devices,
	})
}

func saveDevices(c *gin.Context, config Config) {
	fmt.Println("Saving devices to MongoDB...")

	// Read the request body
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Parse JSON into a slice of maps
	var devices []map[string]interface{}
	if err := json.Unmarshal(body, &devices); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Get the devices collection
	collection := client.Database(config.DatabaseName).Collection(config.CollectionName)

	// Clear the existing collection
	_, err = collection.DeleteMany(context.TODO(), bson.D{}) // Clearing the collection

	if err != nil {
		fmt.Printf("Error deleting existing devices: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update devices"})
		return
	}

	// Insert the updated devices
	for i, device := range devices {
		_, err := collection.InsertOne(context.TODO(), device)
		if err != nil {
			fmt.Printf("Error inserting device %d: %v\n", i, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update devices"}) // Return specific error
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Devices updated successfully"})
}

func main() {
	// Load application configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Construct the MongoDB URI
	mongoURI := fmt.Sprintf("%s:%s", config.MongoDBURL, config.MongoDBPort)
	err = initMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	// Set up the Gin router
	router := gin.Default()

	// Add CORS middleware BEFORE defining routes
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{fmt.Sprintf("%s:%s", config.FrontendURL, config.FrontendPort)}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}

	router.Use(cors.New(corsConfig))

	// Define routes AFTER CORS middleware
	router.GET("/fetch-devices", func(c *gin.Context) {
		fetchAndStoreDevices(c, config)
	})

	router.GET("/devices", func(c *gin.Context) {
		getDevices(c, config)
	})

	router.PUT("/devices", func(c *gin.Context) { // Use PUT for updates
		saveDevices(c, config)
	})

	// Start the server on the configured port
	router.Run(":" + config.ServerPort)
}
