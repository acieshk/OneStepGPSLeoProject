package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"OneStepGPSLeo/api"
	"OneStepGPSLeo/database"
	"OneStepGPSLeo/handlers"
	mockserver "OneStepGPSLeo/mockServer"
	"OneStepGPSLeo/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Define command-line flags
	mockMode := flag.Bool("mock", false, "Run in mock mode")
	mutateChance := flag.Float64("mutateChance", 0.3, "Chance of mutation (0.0 - 1.0)") // Mutation chance flag
	mutateDeviceCount := flag.Int("mutateDevice", 2, "Number of devices to mutate")     // Number of mutations flag
	flag.Parse()
	fmt.Println("Mock mode:", *mockMode)
	fmt.Println("Mutation chance:", *mutateChance)
	fmt.Println("Number of mutations:", *mutateDeviceCount)

	if *mockMode {
		// Get mock server port from config, default to 8081 if not set
		mockServerPort := config.MockServerPort
		if mockServerPort == "" {
			mockServerPort = "8081"
		}
		go mockserver.StartMockServer(config, mockServerPort, 5*time.Second, *mutateChance, *mutateDeviceCount)
		config.APIURL = "http://localhost:8081/api/v1/devices"
		config.APIKey = ""

		fmt.Println("Waiting for mock server to start...") // Indicate waiting
		time.Sleep(2 * time.Second)                        // Give the mock server time to start

		fmt.Println("Mock server started, continuing...") // Indicate continuation
	}

	db, err := database.NewMongoDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	deviceHandlers := handlers.NewDeviceHandlers(config, db)
	userHandlers := handlers.NewUserHandlers(config, db)
	iconHandlers := handlers.NewIconHandlers(config, db)

	lastUpdateTimes := make(map[string]time.Time)
	var updateMutex sync.RWMutex
	lastChecked := time.Now()

	go func() {
		for {
			fmt.Println("Fetching device data from external api")
			api.FetchAndStoreDevices(db, config, &updateMutex, lastUpdateTimes, &lastChecked) // Call from api package

			time.Sleep(time.Duration(config.UpdateInterval) * time.Second) // Correct duration
		}
	}()

	router := gin.Default()
	router.Use(cors.Default())

	apiRoutes := router.Group("/api")
	{
		deviceRoutes := apiRoutes.Group("/devices")
		{
			deviceRoutes.GET("", deviceHandlers.GetDevices)
			deviceRoutes.PUT("/:id", deviceHandlers.UpdateDeviceHandler)
			deviceRoutes.GET("/check-updates", func(c *gin.Context) {
				api.CheckForUpdates(c, db, config, &lastChecked, lastUpdateTimes)
			})
			deviceRoutes.GET("/:id/settings", deviceHandlers.GetDeviceSettingsHandler)
			deviceRoutes.PUT("/:id/settings", deviceHandlers.SaveDeviceSettingsHandler)
			deviceRoutes.DELETE("/refresh", deviceHandlers.RefreshDatabaseHandler)
		}
		userRoutes := apiRoutes.Group("/users")
		{
			userRoutes.GET("/:userId/preferences", userHandlers.GetUserPreferencesHandler)
			userRoutes.POST("/:userId/preferences", userHandlers.SaveUserPreferencesHandler)
		}
	}

	router.POST("/api/devices/:id/icon", iconHandlers.HandleIconUpload)
	router.GET("/api/devices/:id/icon", iconHandlers.GetIconHandler)
	router.Static("/icons", "./icons")

	router.Run(":" + config.ServerPort)

}

func loadConfig(filename string) (models.Config, error) {
	var config models.Config
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}

	// Set default values
	if config.UpdateInterval == 0 {
		config.UpdateInterval = 60
	}
	if config.MockServerPort == "" {
		config.MockServerPort = "8081"
	}
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}

	return config, nil
}
