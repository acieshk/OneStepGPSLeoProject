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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config represents the configuration structure for the application
type Config struct {
	ServerPort           string `json:"server_port"`
	MongoDBURL           string `json:"mongodb_url"`
	MongoDBPort          string `json:"mongodb_port"`
	MongoDBUsername      string `json:"mongodb_username"`
	MongoDBPassword      string `json:"mongodb_password"`
	DatabaseName         string `json:"database_name"`
	DeviceCollectionName string `json:"device_collection_name"`
	UserCollectionName   string `json:"user_collection_name"`
	APIKey               string `json:"api_key"` // API key for external API access
	APIURL               string `json:"api_url"` // Base URL for the API
	FrontendURL          string `json:"frontend_url"`
	FrontendPort         string `json:"frontend_port"`
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
func initMongoDB(config Config) (*mongo.Client, error) {
	// Use config.DatabaseName here to get the database
	mongoURI := fmt.Sprintf("mongodb://%s:%s", config.MongoDBURL, config.MongoDBPort)
	if config.MongoDBUsername != "" && config.MongoDBPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/",
			config.MongoDBUsername,
			config.MongoDBPassword,
			config.MongoDBURL,
			config.MongoDBPort,
		)
	}

	fmt.Println("Connecting to:", mongoURI)
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err) // Return the error immediately
	}

	// Check if there was an error during connection
	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %v", err)
	}

	// Create database if it doesn't exist.
	db := client.Database(config.DatabaseName)

	// Check if the collections exist; if not, create them.
	err = createCollectionIfNotExists(db, config.DeviceCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to create devices collection: %v", err)
	}

	err = createCollectionIfNotExists(db, config.UserCollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user preferences collection: %v", err)
	}

	return client, nil
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
	collection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)
	// Clear existing devices
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		fmt.Printf("Error clearing collection: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing devices"})
		return
	}
	fmt.Printf("Deleted %d existing documents\n", deleteResult.DeletedCount)

	// Store each device directly - Improved error handling
	for i, device := range response.ResultList {
		_, err := collection.InsertOne(context.TODO(), device)
		if err != nil {
			fmt.Printf("Error inserting device %d: %v\n", i, err) // Log the error
			continue                                              // Continue to the next device
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
	collection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)

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

// func saveDevices(c *gin.Context, config Config) {
// 	fmt.Println("Saving devices to MongoDB...")

// 	body, err := ioutil.ReadAll(c.Request.Body)
// 	if err != nil {
// 		log.Printf("Error reading request body: %v", err)                            // Consistent logging
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"}) // Consistent error message
// 		return
// 	}

// 	var devices []map[string]interface{}
// 	if err := json.Unmarshal(body, &devices); err != nil {
// 		log.Printf("Error parsing JSON: %v", err)                                    // Consistent logging
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"}) // Consistent error message
// 		return
// 	}

// 	// Get the devices collection
// 	collection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)

// 	// Clear the existing collection
// 	_, err = collection.DeleteMany(context.TODO(), bson.D{}) // Clearing the collection

// 	if err != nil {
// 		fmt.Printf("Error deleting existing devices: %v\n", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update devices"})
// 		return
// 	}

// 	// Insert the updated devices
// 	for i, device := range devices {
// 		_, err := collection.InsertOne(context.TODO(), device)
// 		if err != nil {
// 			fmt.Printf("Error inserting device %d: %v\n", i, err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update devices"}) // Return specific error
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Devices updated successfully"})
// }

func updateDeviceHandler(c *gin.Context, config Config) {
	deviceIDStr := c.Param("id") // Use Gin's way to get path parameters

	deviceID, err := primitive.ObjectIDFromHex(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"}) // Use Gin's JSON function
		return
	}

	var updatedDevice map[string]interface{} // Use a map to receive updates

	if err := c.ShouldBindJSON(&updatedDevice); err != nil { // Use ShouldBindJSON to handle JSON and other formats
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	//Remove the _id from updatedDevice to avoid conflicts
	delete(updatedDevice, "_id")

	filter := bson.M{"_id": deviceID}
	update := bson.M{"$set": updatedDevice}

	collection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating device"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})

}

// Helper function to create a collection if it doesn't exist
func createCollectionIfNotExists(db *mongo.Database, collectionName string) error {
	list, err := db.ListCollectionNames(context.TODO(), bson.M{"name": collectionName})
	if err != nil {
		return err
	}
	if len(list) == 0 { // Collection doesn't exist
		_ = db.CreateCollection(context.TODO(), collectionName)
		if err != nil {
			return err
		}
		fmt.Printf("Created collection %s\n", collectionName)
	}
	return nil
}

func main() {
	// Load application configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB
	client, err = initMongoDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	// Set up the Gin router
	router := gin.Default()

	// Using cros default in dev environments
	router.Use(cors.Default())

	// Define routes AFTER CORS middleware
	router.GET("/fetch-devices", func(c *gin.Context) {
		fetchAndStoreDevices(c, config)

	})

	router.GET("/devices", func(c *gin.Context) {
		getDevices(c, config)
	})

	router.PUT("/devices/:id", func(c *gin.Context) {
		updateDeviceHandler(c, config) // Pass the config to the handler
	})
	// Start the server on the configured port
	router.Run(":" + config.ServerPort)
}
