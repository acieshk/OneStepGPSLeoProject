/*
Package api provides helper functions for the api package.

This file contains utility functions used for interacting with the MongoDB database.
*/
package api

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// getCurrentDevicesMap retrieves all device IDs from the provided MongoDB collection
// and returns them as a map for efficient existence checks.
// The map keys are the device IDs, and the values are empty structs (for memory efficiency).
// It returns an error if there's an issue during database interaction or data processing.
func getCurrentDevicesMap(collection *mongo.Collection) (map[string]struct{}, error) {
	cursor, err := collection.Find(context.TODO(), bson.M{}) // Find all documents
	if err != nil {
		return nil, fmt.Errorf("failed to find devices: %w", err)
	}
	defer cursor.Close(context.TODO())

	devicesMap := make(map[string]struct{}) // Use empty struct for efficiency

	for cursor.Next(context.TODO()) {
		var device bson.M // Decode into bson.M
		if err := cursor.Decode(&device); err != nil {
			return nil, fmt.Errorf("failed to decode device: %w", err)
		}

		deviceID, ok := device["device_id"].(string)
		if !ok {

			// Log the error or handle it as needed
			fmt.Printf("device_id is missing or not a string for document: %+v", device)

			continue // Skip if device_id isn't a string
		}
		devicesMap[deviceID] = struct{}{} // Use empty struct
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return devicesMap, nil
}
