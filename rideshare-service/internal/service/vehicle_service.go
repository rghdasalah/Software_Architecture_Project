package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/db"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/models"
	"github.com/google/uuid"
)

type VehicleService struct {
	dbManager *db.DBManager
}

func NewVehicleService(dbManager *db.DBManager) *VehicleService {
	return &VehicleService{dbManager: dbManager}
}

// CreateVehicle creates a new vehicle in the database
func (s *VehicleService) CreateVehicle(ctx context.Context, userID uuid.UUID, req *models.CreateVehicleRequest) (*models.Vehicle, error) {
	// Validate request
	if req.Make == "" || req.Model == "" || req.Year <= 0 || 
	   req.Color == "" || req.LicensePlate == "" || req.Capacity <= 0 {
		return nil, errors.New("missing required vehicle fields")
	}

	// Insert vehicle into database
	query := `
		INSERT INTO vehicles (user_id, make, model, year, color, license_plate, capacity, is_active) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, true) 
		RETURNING vehicle_id, user_id, make, model, year, color, license_plate, capacity, is_active, created_at, updated_at
	`

	var vehicle models.Vehicle
	err := s.dbManager.GetPrimary().QueryRowContext(
		ctx,
		query,
		userID,
		req.Make,
		req.Model,
		req.Year,
		req.Color,
		req.LicensePlate,
		req.Capacity,
	).Scan(
		&vehicle.ID,
		&vehicle.UserID,
		&vehicle.Make,
		&vehicle.Model,
		&vehicle.Year,
		&vehicle.Color,
		&vehicle.LicensePlate,
		&vehicle.Capacity,
		&vehicle.IsActive,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &vehicle, nil
}

// GetVehiclesByUserID retrieves all vehicles belonging to a user
func (s *VehicleService) GetVehiclesByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Vehicle, error) {
	query := `
		SELECT vehicle_id, user_id, make, model, year, color, license_plate, capacity, is_active, created_at, updated_at
		FROM vehicles
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vehicles := []*models.Vehicle{}
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.Make,
			&v.Model,
			&v.Year,
			&v.Color,
			&v.LicensePlate,
			&v.Capacity,
			&v.IsActive,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vehicles, nil
}

// GetVehicleByID retrieves a single vehicle by ID
func (s *VehicleService) GetVehicleByID(ctx context.Context, vehicleID uuid.UUID) (*models.Vehicle, error) {
	query := `
		SELECT vehicle_id, user_id, make, model, year, color, license_plate, capacity, is_active, created_at, updated_at
		FROM vehicles
		WHERE vehicle_id = $1 AND is_active = true
	`

	var v models.Vehicle
	err := s.dbManager.GetReplica().QueryRowContext(ctx, query, vehicleID).Scan(
		&v.ID,
		&v.UserID,
		&v.Make,
		&v.Model,
		&v.Year,
		&v.Color,
		&v.LicensePlate,
		&v.Capacity,
		&v.IsActive,
		&v.CreatedAt,
		&v.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("vehicle not found")
	} else if err != nil {
		return nil, err
	}

	return &v, nil
}

// UpdateVehicle updates an existing vehicle
func (s *VehicleService) UpdateVehicle(ctx context.Context, userID, vehicleID uuid.UUID, req *models.CreateVehicleRequest) (*models.Vehicle, error) {
	// First check if the vehicle exists and belongs to the user
	vehicle, err := s.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	if vehicle.UserID != userID {
		return nil, errors.New("vehicle does not belong to user")
	}

	// Update the vehicle
	query := `
		UPDATE vehicles
		SET make = $1, model = $2, year = $3, color = $4, license_plate = $5, capacity = $6, updated_at = NOW()
		WHERE vehicle_id = $7
		RETURNING vehicle_id, user_id, make, model, year, color, license_plate, capacity, is_active, created_at, updated_at
	`

	var updatedVehicle models.Vehicle
	err = s.dbManager.GetPrimary().QueryRowContext(
		ctx,
		query,
		req.Make,
		req.Model,
		req.Year,
		req.Color,
		req.LicensePlate,
		req.Capacity,
		vehicleID,
	).Scan(
		&updatedVehicle.ID,
		&updatedVehicle.UserID,
		&updatedVehicle.Make,
		&updatedVehicle.Model,
		&updatedVehicle.Year,
		&updatedVehicle.Color,
		&updatedVehicle.LicensePlate,
		&updatedVehicle.Capacity,
		&updatedVehicle.IsActive,
		&updatedVehicle.CreatedAt,
		&updatedVehicle.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updatedVehicle, nil
}

// DeleteVehicle soft-deletes a vehicle (sets is_active to false)
func (s *VehicleService) DeleteVehicle(ctx context.Context, userID, vehicleID uuid.UUID) error {
	// Check if the vehicle exists and belongs to the user
	vehicle, err := s.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		return err
	}

	if vehicle.UserID != userID {
		return errors.New("vehicle does not belong to user")
	}

	// Soft delete the vehicle
	query := `
		UPDATE vehicles
		SET is_active = false, updated_at = NOW()
		WHERE vehicle_id = $1
	`

	_, err = s.dbManager.GetPrimary().ExecContext(ctx, query, vehicleID)
	return err
}