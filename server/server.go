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

type UserPreferences struct {
	UserID       string `bson:"user_id" json:"userId"`             // Changed to userId to match request
	DistanceUnit string `bson:"distance_unit" json:"distanceUnit"` // Match casing from frontend
	Layout       string `bson:"layout" json:"layout"`              // Match casing from frontend

}

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

// Handlers for passing dependencies
type Handlers struct {
	Config                    Config
	DevicesCollection         *mongo.Collection
	UserPreferencesCollection *mongo.Collection
	DBClient                  *mongo.Client
}

// Global variable to hold the MongoDB client
var (
	client   *mongo.Client
	handlers *Handlers
)

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

func fetchAndStoreDevices(c *gin.Context, h *Handlers) {
	fmt.Println("Starting fetchAndStoreDevices...")
	apiURL := fmt.Sprintf("%s%s", h.Config.APIURL, h.Config.APIKey)
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
	collection := client.Database(h.Config.DatabaseName).Collection(h.Config.DeviceCollectionName)
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

func getDevices(c *gin.Context, h *Handlers) {
	fmt.Println("Getting devices from MongoDB...")
	collection := client.Database(h.Config.DatabaseName).Collection(h.Config.DeviceCollectionName)

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

func updateDeviceHandler(c *gin.Context, h *Handlers) {
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

	collection := h.DBClient.Database(h.Config.DatabaseName).Collection(h.Config.DeviceCollectionName)

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

	// Create the icon directory if it doesn't exist
	iconDir := "./icons"
	if err := os.MkdirAll(iconDir, os.ModePerm); err != nil {

		return
	}

	filename := fmt.Sprintf("%s.png", deviceID.Hex()) // Use device ID as filename + extension

	filepath := filepath.Join(iconDir, filename)

	// Save file locally
	outFile, err := os.Create(filepath)

	defer outFile.Close()

	_, err = io.Copy(outFile, file)

	iconURL := fmt.Sprintf("%s:%s/icons/%s", h.Config.FrontendURL, h.Config.FrontendPort, filename)

	if err := updateDeviceIconURL(h, deviceID, iconURL); err != nil {
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
	update := bson.M{"$set": bson.M{"iconUrl": iconURL}} // Assuming "icon_url" is the field in MongoDB

	_, err := h.DevicesCollection.UpdateOne(context.TODO(), filter, update) // Use the correct collection

	if err != nil {
		return fmt.Errorf("failed to update iconURL %v", err)
	}

	return nil
}

func getUserPreferencesHandler(c *gin.Context, h *Handlers) {
	userId := c.Param("userId")

	var prefs UserPreferences
	err := h.UserPreferencesCollection.FindOne(context.TODO(), bson.M{"user_id": userId}).Decode(&prefs) // Use "user_id"
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default preferences if not found
			prefs = UserPreferences{
				UserID:       userId, // Or generate a new ID if needed
				DistanceUnit: "km",
				Layout:       "horizontal",
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve preferences"})
			return
		}
	}

	c.JSON(http.StatusOK, prefs) // Return preferences
}

func saveUserPreferencesHandler(c *gin.Context, h *Handlers) {
	var prefs UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Use upsert to create or update preferences
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": prefs.UserID} // Use the correct field name in the filter!
	update := bson.M{"$set": prefs}           // Use $set to update fields

	// Use the correct collection from your Handlers (h) and a context
	result, err := h.UserPreferencesCollection.UpdateOne(context.TODO(), filter, update, opts) // Corrected context usage
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save preferences %v", err)}) // Return error to frontend
		return
	}

	// Log details for debugging
	if result.UpsertedCount > 0 {
		fmt.Println("Inserted a single document: ", result.UpsertedID)
	} else {
		fmt.Println("Matched", result.MatchedCount, "documents and updated", result.ModifiedCount, "documents.")
	}

	c.JSON(http.StatusOK, prefs) // Respond with the updated preferences
}

// If icon exists, return the path
// Otherwise, return ""
func getIconHandler(c *gin.Context) {
	deviceId := c.Param("id") // Use Gin's param method

	iconPath := filepath.Join("icons", deviceId+".png")

	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		c.String(http.StatusOK, "") // Return empty string with 200 OK if not found (Gin's way)
		return
	}

	// For serving static files, Gin's File function is often cleaner:
	c.File(iconPath) // Gin automatically handles Content-Type
}

func main() {
	// Load application configuration
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB client *before* creating collections or handlers
	client, err = initMongoDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Initialize collections (after initializing MongoDB!)
	devicesCollection := client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)
	userPreferencesCollection := client.Database(config.DatabaseName).Collection(config.UserCollectionName)

	// Initialize handlers (after MongoDB and collections are initialized!)

	handlers = &Handlers{
		Config:                    config,
		DevicesCollection:         devicesCollection,
		UserPreferencesCollection: userPreferencesCollection,
		DBClient:                  client,
	}

	// Set up the Gin router
	router := gin.Default()

	// CORS Configuration (Simplified for development)
	router.Use(cors.Default())

	// Define routes
	router.POST("/fetch-devices", func(c *gin.Context) { fetchAndStoreDevices(c, handlers) })
	router.GET("/devices", func(c *gin.Context) { getDevices(c, handlers) })
	router.PUT("/devices/:id", func(c *gin.Context) { updateDeviceHandler(c, handlers) })
	router.POST("/devices/:id/icon", func(c *gin.Context) { handleIconUpload(c, handlers) })
	router.GET("/user-preferences/:userId", func(c *gin.Context) { getUserPreferencesHandler(c, handlers) })
	router.POST("/user-preferences", func(c *gin.Context) { saveUserPreferencesHandler(c, handlers) })
	router.GET("/getIcon/:id", getIconHandler) // Use Gin's router for getIcon

	router.Static("/icons", "./icons") // Serve static files.

	// Start the server *after* all initialization
	router.Run(":" + config.ServerPort)

}
