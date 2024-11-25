package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	_ "image/gif"  // Decode GIF
	_ "image/jpeg" // Decode JPEG
	_ "image/png"  // Decode PNG
	"path/filepath"

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

// Define the Handlers type (if not already defined)
type Handlers struct {
	Config                    Config
	DevicesCollection         *mongo.Collection
	UserPreferencesCollection *mongo.Collection
	DBClient                  *mongo.Client
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
	// Check if the database was actually created (if it didn't already exist)
	// Get a list of database names
	dbNames, err := client.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list database names: %w", err) // Wrap the error
	}

	dbExists := false
	for _, name := range dbNames {
		if name == config.DatabaseName {
			dbExists = true
			break
		}

	}

	if !dbExists {
		fmt.Printf("Created database: %s\n", config.DatabaseName)

	}

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

func handleIconUpload(c *gin.Context, h *Handlers) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve uploaded file"})
		return
	}
	defer file.Close()

	// Validate file type and size
	if err := validateImageFile(file, header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reset file pointer after validation
	_, err = file.Seek(0, io.SeekStart) // Resetting after reading the file header
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to reset file pointer %v", err)})
		return
	}

	iconsDir := "./device-icons" // Path for storing icons
	if err := os.MkdirAll(iconsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create icon directory"})
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := primitive.ObjectIDFromHex(deviceIDStr)

	// Create the icon directory if it doesn't exist.  Use os.MkdirAll for nested directories.
	iconDir := "./icons" // Top-level icon directory
	if err := os.MkdirAll(iconDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create icon directory: %w", err)})
		return
	}

	filename := fmt.Sprintf("%s.png", deviceID.Hex()) //  Use device ID as filename + extension
	filepath := filepath.Join(iconDir, filename)

	// Save file locally
	outFile, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create icon file"})
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save icon file"})
		return
	}

	iconURL := fmt.Sprintf("%s:%s/icons/%s", h.Config.FrontendURL, h.Config.FrontendPort, filename)

	if err := updateDeviceIconURL(h, deviceID, iconURL); err != nil { // Pass *Handlers
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device icon URL"})

		return

	}

	c.JSON(http.StatusOK, gin.H{"iconURL": iconURL})
}

func validateImageFile(file io.Reader, header *multipart.FileHeader) error {

	// Decode to get image dimensions and validate it's an image
	_, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %v", err)
	}

	// ... other validation if needed (e.g., file size)

	return nil
}

func saveIconFile(file multipart.File, filename string, h *Handlers) error {

	// Create the device-icons directory if it doesn't exist
	if err := os.MkdirAll("./device-icons", os.ModePerm); err != nil { // corrected directory path
		return fmt.Errorf("failed to create device-icons directory: %w", err)
	}

	out, err := os.Create(filepath.Join("./device-icons", filename)) // Save in device-icons
	if err != nil {
		return fmt.Errorf("failed to create icon file: %w", err)
	}

	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to copy icon file: %w", err)

	}

	return nil
}

func updateDeviceIconURL(h *Handlers, deviceID primitive.ObjectID, iconURL string) error {
	// ... your logic to update the iconURL in the database (similar to updateDeviceHandler)
	filter := bson.M{"_id": deviceID}
	update := bson.M{"$set": bson.M{"icon_url": iconURL}} // Assuming "icon_url" is the field in MongoDB

	_, err := h.DevicesCollection.UpdateOne(context.TODO(), filter, update) // Use the correct collection

	if err != nil {
		return fmt.Errorf("failed to update iconURL %v", err)
	}

	return nil
}

var handlers *Handlers

func main() {
	// Load application configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB client and collections
	client, err = initMongoDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err) // Correct: log and exit on config error
	}
	// Initialize collections
	devicesCollection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)
	userPreferencesCollection := client.Database(config.DatabaseName).Collection(config.UserCollectionName)
	// Initialize Handlers
	handlers := &Handlers{
		Config:                    config,
		DevicesCollection:         devicesCollection,
		UserPreferencesCollection: userPreferencesCollection,
		DBClient:                  client,
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

	router.POST("/devices/:id/icon", func(c *gin.Context) {
		handleIconUpload(c, handlers) // Pass handlers, not config
	})
	router.Static("/icons", "./icons")

	// Start the server on the configured port
	router.Run(":" + config.ServerPort)

}
