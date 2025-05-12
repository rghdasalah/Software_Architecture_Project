package repository

import (
    "context"
    "log"
    "location-service/models"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// LocationRepository interface defines the methods for interacting with location data
type LocationRepository interface {
    Save(ctx context.Context, update models.LocationUpdate) error
    GetLatest(ctx context.Context, rideID int) (*models.LocationUpdate, error)
}

type locationRepository struct {
    collection *mongo.Collection
}

func NewLocationRepository(client *mongo.Client) LocationRepository {
    collection := client.Database("rideshare").Collection("location_updates")
    return &locationRepository{collection: collection}
}

func (r *locationRepository) Save(ctx context.Context, update models.LocationUpdate) error {
    _, err := r.collection.InsertOne(ctx, update)
    if err != nil {
        log.Printf("Failed to save location update: %v", err)
        return err
    }
    return nil
}

func (r *locationRepository) GetLatest(ctx context.Context, rideID int) (*models.LocationUpdate, error) {
    var update models.LocationUpdate
    opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
    err := r.collection.FindOne(ctx, bson.M{"ride_id": rideID}, opts).Decode(&update)
    if err != nil {
        log.Printf("Failed to get latest location for ride %d: %v", rideID, err)
        return nil, err
    }
    return &update, nil
}