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
	Client               *mongo.Client
	DatabaseName         string
	Config               models.Config // or your config package's type
	DeviceCollectionName string
	UserCollectionName   string
}

func NewMongoDB(cfg models.Config) (*MongoDB, error) {
	mongoURI := fmt.Sprintf("mongodb://%s:%s", cfg.MongoDBURL, cfg.MongoDBPort)
	if cfg.MongoDBUsername != "" && cfg.MongoDBPassword != "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/",
			cfg.MongoDBUsername,
			cfg.MongoDBPassword,
			cfg.MongoDBURL,
			cfg.MongoDBPort,
		)
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping failed: %v", err)
	}

	db := client.Database(cfg.DatabaseName)

	// Check if collections exist and create if they don't. Wrap errors!
	if err := createCollectionIfNotExists(db, cfg.DeviceCollectionName); err != nil {
		return nil, fmt.Errorf("failed to create devices collection: %w", err) // Wrap errors for better context
	}

	if err := createCollectionIfNotExists(db, cfg.UserCollectionName); err != nil {
		return nil, fmt.Errorf("failed to create users collection: %w", err) // Wrap errors
	}

	return &MongoDB{
		Client:               client,
		DatabaseName:         cfg.DatabaseName,
		DeviceCollectionName: cfg.DeviceCollectionName,
		UserCollectionName:   cfg.UserCollectionName,
		Config:               cfg}, nil
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
			// You can either return an error, create a new document, or return a default value.
			return models.UserPreferences{}, fmt.Errorf("user preferences not found: %w", err)
		} else {
			// Handle other errors that might occur during the query.
			return models.UserPreferences{}, fmt.Errorf("failed to get user preferences: %w", err)
		}
	}

	return prefs, nil
}

func (db *MongoDB) SaveUserPreferences(prefs models.UserPreferences) error {

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": prefs.UserID}
	update := bson.M{"$set": prefs}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := db.Client.Database(db.DatabaseName).Collection(db.UserCollectionName).UpdateOne(ctx, filter, update, opts) // Use the client from the struct
	return err

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
