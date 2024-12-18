package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"OneStepGPSLeo/models" // Import your models package

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client                 *mongo.Client
	DatabaseName           string
	Config                 models.Config
	DeviceCollectionName   string
	UserCollectionName     string
	SettingsCollectionName string
}

func NewMongoDB(cfg models.Config) (*MongoDB, error) {
	// Setup MongoDB URI
	mongoURI := fmt.Sprintf("mongodb://%s:%s", cfg.MongoDBURL, cfg.MongoDBPort)
	if cfg.MongoDBUsername != "" && cfg.MongoDBPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/",
			cfg.MongoDBUsername,
			cfg.MongoDBPassword,
			cfg.MongoDBURL,
			cfg.MongoDBPort,
		)
	}

	// Create context with timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %v", err)
	}

	// Create database if it doesn't exist
	if err := createDatabaseIfNotExists(ctx, client, cfg.DatabaseName); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	db := client.Database(cfg.DatabaseName)
	// Check if collections exist and create if they don't
	if err := createCollectionIfNotExists(db, cfg.DeviceCollectionName); err != nil {
		return nil, fmt.Errorf("failed to create devices collection: %w", err)
	}

	if err := createCollectionIfNotExists(db, cfg.UserCollectionName); err != nil {
		return nil, fmt.Errorf("failed to create users collection: %w", err)
	}

	if err := createCollectionIfNotExists(db, cfg.SettingsCollectionName); err != nil {
		return nil, fmt.Errorf("failed to create settings collection: %w", err)
	}

	return &MongoDB{
		Client:                 client,
		DatabaseName:           cfg.DatabaseName,
		DeviceCollectionName:   cfg.DeviceCollectionName,
		UserCollectionName:     cfg.UserCollectionName,
		SettingsCollectionName: cfg.SettingsCollectionName,
		Config:                 cfg,
	}, nil
}

func createDatabaseIfNotExists(ctx context.Context, client *mongo.Client, databaseName string) error {

	list, err := client.ListDatabaseNames(ctx, bson.M{}) //List the existing database
	if err != nil {
		return fmt.Errorf("failed to list existing databases: %w", err)
	}
	for _, dbName := range list {
		if dbName == databaseName {
			return nil //Return nil if database exists to prevent creation.
		}
	}

	// If the database does not exist, create it here.
	if err := client.Database(databaseName).CreateCollection(ctx, "init"); err != nil {
		return fmt.Errorf("failed to create database %s: %w", databaseName, err)

	} else {
		// Delete the "init" collection immediately. The "init" collection is used to ensure
		// that the named database is created if it does not exist.
		client.Database(databaseName).Collection("init").Drop(ctx)
	}

	return nil

}

func createCollectionIfNotExists(db *mongo.Database, collectionName string) error {
	list, err := db.ListCollectionNames(context.TODO(), bson.M{"name": collectionName})
	if err != nil {
		return fmt.Errorf("failed to list collection names: %w", err) // Wrap error for context
	}
	if len(list) == 0 {
		if err := db.CreateCollection(context.TODO(), collectionName); err != nil { // Check error during creation
			return fmt.Errorf("failed to create collection: %w", err) // Wrap the error
		}
		fmt.Printf("Created collection %s\n", collectionName)
	}
	return nil
}

func (db *MongoDB) GetDevices() ([]bson.M, error) {
	collection := db.Client.Database(db.DatabaseName).Collection(db.DeviceCollectionName)

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find devices: %w", err)
	}
	defer cursor.Close(context.TODO())

	var devices []bson.M // Use []bson.M to store the results
	if err = cursor.All(context.TODO(), &devices); err != nil {
		return nil, fmt.Errorf("failed to decode devices: %w", err)
	}

	return devices, nil
}

func (db *MongoDB) UpdateDevice(deviceID primitive.ObjectID, updatedDevice map[string]interface{}, deviceVersion int) error {
	filter := bson.M{"_id": deviceID, "version": deviceVersion}

	updateMap := make(map[string]interface{})
	for k, v := range updatedDevice {
		updateMap[k] = v
	}
	delete(updateMap, "_id") // Don't allow _id to be changed by a client
	delete(updateMap, "version")

	update := bson.M{"$set": updateMap, "$inc": bson.M{"version": 1}} //Increment version atomically after update

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := db.Client.Database(db.DatabaseName).Collection(db.DeviceCollectionName).FindOneAndUpdate(context.TODO(), filter, update, opts)

	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return fmt.Errorf("outdated device version") // Specific error for version mismatch
		}
		return fmt.Errorf("failed to update device: %w", result.Err())
	}

	return nil
}

func (db *MongoDB) UpdateDeviceIconURL(deviceID primitive.ObjectID, iconURL string) error {
	filter := bson.M{"_id": deviceID}
	update := bson.M{"$set": bson.M{"iconUrl": iconURL}} // Use the field name "iconUrl"
	collection := db.Client.Database(db.DatabaseName).Collection(db.DeviceCollectionName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update iconURL: %v", err) //Better error message
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("device not found or iconURL not updated") // Informative error for not found or failed update
	}
	return nil
}

func (db *MongoDB) GetUserPreferences(userID string) (models.UserPreferences, error) {
	// ... similar to GetDevices, query the UserPreferencesCollection and return a models.UserPreferences
	collection := db.Client.Database(db.DatabaseName).Collection(db.UserCollectionName)

	var prefs models.UserPreferences
	err := collection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&prefs)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Handle the case where no matching document is found.
			return models.UserPreferences{}, fmt.Errorf("user preferences not found: %w", err)
		} else {
			// Handle other errors that might occur during the query.
			return models.UserPreferences{}, fmt.Errorf("failed to get user preferences: %w", err)
		}
	}

	return prefs, nil
}

var (
	ErrPreferencesNotFound = errors.New("user preferences not found") //Custom error
	ErrOutdatedVersion     = errors.New("outdated preferences version")
)

func (db *MongoDB) SaveUserPreferences(prefs models.UserPreferences) (models.UserPreferences, error) {
	collection := db.Client.Database(db.DatabaseName).Collection(db.UserCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add version check to filter
	filter := bson.M{
		"user_id": prefs.UserID,
		"version": prefs.Version, // Only update if version matches
	}

	// Remove version from the prefs object for $set
	prefsWithoutVersion := bson.M{
		"user_id":           prefs.UserID,
		"device_list_width": prefs.DeviceListWidth,
		"unit":              prefs.Unit,
	}

	opts := options.FindOneAndUpdate().SetUpsert(false).SetReturnDocument(options.After)

	update := bson.M{
		"$set": prefsWithoutVersion,
		"$inc": bson.M{"version": 1},
	}

	result := collection.FindOneAndUpdate(ctx, filter, update, opts)

	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// First, check if the document exists with a different version
			existingDoc := collection.FindOne(ctx, bson.M{"user_id": prefs.UserID})
			if existingDoc.Err() == nil {
				var currentPrefs models.UserPreferences
				if err := existingDoc.Decode(&currentPrefs); err == nil {
					// Document exists but version doesn't match
					return currentPrefs, ErrOutdatedVersion
				}
			}
			// If document doesn't exist at all, create new
			prefs.Version = 1
			_, err := collection.InsertOne(ctx, prefs)
			if err != nil {
				return models.UserPreferences{}, fmt.Errorf("failed to create new preferences: %w", err)
			}
			return prefs, nil
		}
		return models.UserPreferences{}, fmt.Errorf("failed to save preferences: %w", result.Err())
	}

	var updatedPrefs models.UserPreferences
	if err := result.Decode(&updatedPrefs); err != nil {
		return models.UserPreferences{}, fmt.Errorf("failed to decode preferences: %w", err)
	}

	return updatedPrefs, nil
}

func (db *MongoDB) ClearCollections() error {
	_, err := db.Client.Database(db.DatabaseName).Collection(db.DeviceCollectionName).DeleteMany(context.TODO(), bson.M{})

	if err != nil {

		return fmt.Errorf("failed to clear devices collection: %w", err)
	}

	// Clear user preferences collection
	_, err = db.Client.Database(db.DatabaseName).Collection(db.UserCollectionName).DeleteMany(context.TODO(), bson.M{}) // Corrected collection name

	if err != nil {
		return fmt.Errorf("failed to clear user preferences collection: %w", err)

	}
	return nil // Return nil if successful
}

func (db *MongoDB) GetDeviceSettings(deviceID string) (models.DeviceSettings, error) {
	var settings models.DeviceSettings
	filter := bson.M{"device_id": deviceID}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.Client.Database(db.DatabaseName).Collection(db.SettingsCollectionName).FindOne(ctx, filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Handle "not found" by creating a new document if needed.
			currentTime := time.Now().Format(time.RFC3339)
			settings = models.DeviceSettings{
				DeviceID:              deviceID,
				IconURL:               "",
				Version:               1,
				UpdatedAt:             currentTime,
				BeginMovingSpeed:      models.Speed{Value: 0, Unit: "mph", Display: "0 mph"},
				BeginStoppedSpeed:     models.Speed{Value: 0, Unit: "mph", Display: "0 mph"},
				MaxDriftDistance:      models.Speed{Value: 350, Unit: "m", Display: "350 m"},
				MinNumSatellites:      8,
				IgnoreUnsetMinNumSats: true,
				MaxHdop:               3.5,
				DriveTimeout:          models.Speed{Value: 1800, Unit: "s", Display: "30m"},
				StopTimeout:           models.Speed{Value: 14400, Unit: "s", Display: "4h"},
				OfflineTimeout:        models.Speed{Value: 3900, Unit: "s", Display: "1h 5m"},
				HistoryCalcDuration:   models.Speed{Value: 86400, Unit: "s", Display: "24h"},
				FuelConsumption: models.FuelConsumption{
					CalculationMethod: "fuel_sensor",
					Measurement:       "mpg",
					FuelType:          "",
					FuelCost:          0,
					FuelEconomy:       0,
				},
				InitialDevicePointDeleteCutoffTime: "2024-06-21T17:45:09.284403Z",
				EngineHoursCounterConfig:           "best",
				UseV3EngineHours:                   true,
				HistoryRetentionDays:               1095,
				HarshEventMinSpeed:                 models.Speed{Value: 0, Unit: "mph", Display: "0 mph"},
			}
			if _, err := db.Client.Database(db.DatabaseName).Collection(db.SettingsCollectionName).InsertOne(context.TODO(), settings); err != nil {
				return models.DeviceSettings{}, fmt.Errorf("error creating default device settings: %w", err) // Return error if default creation fails.
			}
			return settings, nil // Return newly created settings.

		}
		return models.DeviceSettings{}, fmt.Errorf("failed to get device settings: %w", err)

	}
	return settings, nil
}

func (db *MongoDB) SaveDeviceSettings(settings models.DeviceSettings) (models.DeviceSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set current timestamp
	settings.UpdatedAt = time.Now().Format(time.RFC3339)

	// Filter to match only device ID
	filter := bson.M{"device_id": settings.DeviceID}
	collection := db.Client.Database(db.DatabaseName).Collection(db.SettingsCollectionName)

	// Fetch existing settings for comparison. Add error handling for FindOne operation if needed.
	var existingSettings models.DeviceSettings
	err := collection.FindOne(ctx, filter).Decode(&existingSettings)

	if err == nil { // Existing settings found
		if settings.Version == 0 {
			settings.Version = 1
		} else if settings.Version == existingSettings.Version { // Versions match - check timestamp
			incomingTimestamp, err := time.Parse(time.RFC3339, settings.UpdatedAt)
			if err != nil {
				return models.DeviceSettings{}, fmt.Errorf("invalid incoming timestamp: %w", err)
			}

			currentTimestamp, err := time.Parse(time.RFC3339, existingSettings.UpdatedAt)
			if err != nil {
				return models.DeviceSettings{}, fmt.Errorf("invalid current timestamp: %w", err)
			}

			if incomingTimestamp.Before(currentTimestamp) || incomingTimestamp.Equal(currentTimestamp) {
				return existingSettings, nil // Return existing settings without updating - no changes
			}
		}

		// Construct update operation - include all relevant fields
		update := bson.M{
			"$inc": bson.M{"version": 1},
			"$set": bson.M{
				"updated_at":                              settings.UpdatedAt,
				"iconUrl":                                 settings.IconURL,
				"begin_moving_speed":                      settings.BeginMovingSpeed,
				"begin_stopped_speed":                     settings.BeginStoppedSpeed,
				"max_drift_distance":                      settings.MaxDriftDistance,
				"min_num_satellites":                      settings.MinNumSatellites,
				"ignore_unset_min_num_sats":               settings.IgnoreUnsetMinNumSats,
				"max_hdop":                                settings.MaxHdop,
				"drive_timeout":                           settings.DriveTimeout,
				"stop_timeout":                            settings.StopTimeout,
				"offline_timeout":                         settings.OfflineTimeout,
				"history_calc_duration":                   settings.HistoryCalcDuration,
				"fuel_consumption":                        settings.FuelConsumption,
				"initial_device_point_delete_cutoff_time": settings.InitialDevicePointDeleteCutoffTime,
				"engine_hours_counter_config":             settings.EngineHoursCounterConfig,
				"use_v3_engine_hours":                     settings.UseV3EngineHours,
				"history_retention_days":                  settings.HistoryRetentionDays,
				"harsh_event_min_speed":                   settings.HarshEventMinSpeed,
			},
		}

		// FindOneAndUpdate handles both updates and inserts
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After) //Return updated document
		result := collection.FindOneAndUpdate(ctx, filter, update, opts)

		if result.Err() != nil { //Handle error from FindOneAndUpdate
			if errors.Is(result.Err(), mongo.ErrNoDocuments) {
				//Handle no document found for whatever reason

			}
			//Handle other database errors
		}

		// Decode updated document and return
		var updatedSettings models.DeviceSettings //Return error if decoding failed.
		if err := result.Decode(&updatedSettings); err != nil {
			return updatedSettings, fmt.Errorf("failed to decode updated settings: %w", err) //Return descriptive error message
		}

		return updatedSettings, nil //Return updated settings
	} else if err == mongo.ErrNoDocuments { //If not found, create new settings.

		settings.Version = 1 // Initialize version for new settings
		_, err := collection.InsertOne(ctx, settings)
		if err != nil {
			return models.DeviceSettings{}, fmt.Errorf("failed to insert new device settings: %w", err) //Handle the insert error
		}
		return settings, nil //Return newly inserted settings

	} else { // Some other error occurred when trying to fetch settings.

		return models.DeviceSettings{}, fmt.Errorf("failed to get existing settings: %w", err) //Handle or log error

	}

	return settings, nil // Return updated/inserted settings
}

func (db *MongoDB) GetIconMap() (map[string]string, error) {
	collection := db.Client.Database(db.DatabaseName).Collection(db.SettingsCollectionName) //Correct collection name
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)                //Add context with timeout
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"device_id": 1, "iconUrl": 1})) // Project only necessary fields for efficiency
	if err != nil {
		return nil, fmt.Errorf("failed to find device settings: %w", err)
	}
	defer cursor.Close(ctx) // Close cursor using ctx

	iconMap := make(map[string]string)
	for cursor.Next(ctx) { //Use ctx for iteration
		var setting struct { // Decode into anonymous struct with relevant fields.
			DeviceID string `bson:"device_id"`
			IconURL  string `bson:"iconUrl"`
		}
		if err := cursor.Decode(&setting); err != nil { //Handle error during document decoding. Add logging here if needed.
			return nil, fmt.Errorf("failed to decode device setting: %w", err)

		}
		iconMap[setting.DeviceID] = setting.IconURL // Use DeviceID, no _id from MongoDB

	}

	if err := cursor.Err(); err != nil { // Check for cursor errors
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return iconMap, nil
}
