package service

import (
    "context"
    "testing"
    "time"
    "location-service/models"
    "location-service/repository" // Used via MockRepository implementing LocationRepository
    "location-service/queue"     // Used via MockQueue implementing LocationQueue

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockRepository implements the LocationRepository interface
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, update models.LocationUpdate) error {
    args := m.Called(ctx, update)
    return args.Error(0)
}

func (m *MockRepository) GetLatest(ctx context.Context, rideID int) (*models.LocationUpdate, error) {
    args := m.Called(ctx, rideID)
    return args.Get(0).(*models.LocationUpdate), args.Error(1)
}

// MockQueue implements the LocationQueue interface
type MockQueue struct {
    mock.Mock
}

func (m *MockQueue) Consume(handler func(models.LocationUpdate)) error {
    args := m.Called(handler)
    return args.Error(0)
}

func (m *MockQueue) Publish(update models.LocationUpdate) error {
    args := m.Called(update)
    return args.Error(0)
}

func TestLocationService(t *testing.T) {
    // Create mocks
    mockRepo := new(MockRepository)
    mockQueue := new(MockQueue)

    // Explicitly use the interfaces to satisfy the compiler
    var _ repository.LocationRepository = mockRepo // Type assertion to use the import
    var _ queue.LocationQueue = mockQueue          // Type assertion to use the import

    // Initialize service
    svc := NewLocationService(mockRepo, mockQueue)

    // Test Start
    update := models.LocationUpdate{
        RideID:    456,
        Latitude:  30.0150,
        Longitude: 31.0000,
        Timestamp: time.Now(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }

    // Capture the handler function passed to Consume
    var capturedHandler func(models.LocationUpdate)
    mockQueue.On("Consume", mock.Anything).Run(func(args mock.Arguments) {
        capturedHandler = args.Get(0).(func(models.LocationUpdate))
    }).Return(nil)

    // Expect the Save call that will happen when the handler is invoked
    mockRepo.On("Save", context.Background(), mock.MatchedBy(func(u models.LocationUpdate) bool {
        return u.RideID == 456 && u.Latitude == 30.0150 && u.Longitude == 31.0000
    })).Return(nil)

    // Call Start, which will call Consume and allow us to capture the handler
    err := svc.Start(context.Background())
    assert.NoError(t, err)

    // Simulate message consumption by invoking the captured handler with the update
    if capturedHandler != nil {
        capturedHandler(update)
    }

    mockQueue.AssertExpectations(t)
    mockRepo.AssertExpectations(t)

    // Test GetLatestLocation
    expected := &models.LocationUpdate{RideID: 456, Latitude: 30.0150, Longitude: 31.0000}
    mockRepo.On("GetLatest", context.Background(), 456).Return(expected, nil)
    result, err := svc.GetLatestLocation(456)
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}