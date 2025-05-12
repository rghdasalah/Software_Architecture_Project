package service

import (
    "context"
    "location-service/models"
    "location-service/repository"
    "location-service/queue"
)

// LocationServiceInterface defines the methods for the Location Service
type LocationServiceInterface interface {
    Start(ctx context.Context) error
    GetLatestLocation(rideID int) (*models.LocationUpdate, error)
}

type LocationService struct {
    repo  repository.LocationRepository // Use interface
    queue queue.LocationQueue          // Use interface
}

func NewLocationService(repo repository.LocationRepository, queue queue.LocationQueue) LocationServiceInterface {
    return &LocationService{repo: repo, queue: queue}
}

func (s *LocationService) Start(ctx context.Context) error {
    return s.queue.Consume(func(update models.LocationUpdate) {
        if err := s.repo.Save(context.Background(), update); err != nil {
            // Log error but continue to next message
            // In a production system, you might want to retry or send to a dead-letter queue
        }
    })
}

func (s *LocationService) GetLatestLocation(rideID int) (*models.LocationUpdate, error) {
    return s.repo.GetLatest(context.Background(), rideID)
}