package api

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "location-service/models"
    "location-service/service"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockLocationService mocks the LocationServiceInterface
type MockLocationService struct {
    mock.Mock
}

func (m *MockLocationService) Start(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockLocationService) GetLatestLocation(rideID int) (*models.LocationUpdate, error) {
    args := m.Called(rideID)
    return args.Get(0).(*models.LocationUpdate), args.Error(1)
}

func TestGetLatestLocation(t *testing.T) {
    // Explicitly use the service package to satisfy the compiler
    var _ service.LocationServiceInterface = new(MockLocationService)

    mockSvc := new(MockLocationService)
    handler := NewHandler(mockSvc)

    // Test case 1: Successful retrieval
    t.Run("Success", func(t *testing.T) {
        expected := &models.LocationUpdate{
            RideID:    123,
            Latitude:  40.7128,
            Longitude: -74.0060,
            Timestamp: time.Now(),
            ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
        }
        mockSvc.On("GetLatestLocation", 123).Return(expected, nil)

        req := httptest.NewRequest("GET", "/location/123", nil)
        w := httptest.NewRecorder()
        handler.GetLatestLocation(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        var result models.LocationUpdate
        err := json.NewDecoder(w.Body).Decode(&result)
        assert.NoError(t, err)
        assert.Equal(t, expected.RideID, result.RideID)
        assert.Equal(t, expected.Latitude, result.Latitude)
        assert.Equal(t, expected.Longitude, result.Longitude)
        mockSvc.AssertExpectations(t)
    })

    // Test case 2: Invalid rideID
    t.Run("InvalidRideID", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/location/invalid", nil)
        w := httptest.NewRecorder()
        handler.GetLatestLocation(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        assert.Contains(t, w.Body.String(), "Invalid rideID")
    })

    // Test case 3: Location not found
    t.Run("NotFound", func(t *testing.T) {
        mockSvc.On("GetLatestLocation", 456).Return((*models.LocationUpdate)(nil), nil)

        req := httptest.NewRequest("GET", "/location/456", nil)
        w := httptest.NewRecorder()
        handler.GetLatestLocation(w, req)

        assert.Equal(t, http.StatusNotFound, w.Code)
        assert.Contains(t, w.Body.String(), "Location not found")
        mockSvc.AssertExpectations(t)
    })
}