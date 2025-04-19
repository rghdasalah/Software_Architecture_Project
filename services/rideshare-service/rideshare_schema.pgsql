--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4 (Debian 17.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.4 (Debian 17.4-1.pgdg120+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- PostgreSQL database schema for Rideshare Service
--

-- Create an extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enums
CREATE TYPE user_role AS ENUM ('rider', 'driver', 'admin');
CREATE TYPE ride_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled');
CREATE TYPE request_status AS ENUM ('pending', 'accepted', 'rejected', 'cancelled');

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    role user_role NOT NULL DEFAULT 'rider',
    profile_picture_url VARCHAR(255),
    date_of_birth DATE NOT NULL,
    bio TEXT,
    average_rating DECIMAL(3,2) DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Create vehicles table
CREATE TABLE IF NOT EXISTS vehicles (
    vehicle_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(50) NOT NULL,
    license_plate VARCHAR(20) NOT NULL,
    capacity INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_year CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1)
);

-- Create rides table
CREATE TABLE IF NOT EXISTS rides (
    ride_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_id UUID NOT NULL REFERENCES users(user_id),
    vehicle_id UUID REFERENCES vehicles(vehicle_id),
    origin_address TEXT NOT NULL,
    origin_latitude DECIMAL(9,6) NOT NULL,
    origin_longitude DECIMAL(9,6) NOT NULL,
    destination_address TEXT NOT NULL,
    destination_latitude DECIMAL(9,6) NOT NULL,
    destination_longitude DECIMAL(9,6) NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    estimated_arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
    max_passengers INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    price_per_seat DECIMAL(10,2),
    route_polyline TEXT,
    status ride_status DEFAULT 'scheduled',
    description TEXT,
    luggage_capacity TEXT,
    is_pets_allowed BOOLEAN DEFAULT FALSE,
    is_smoking_allowed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_seats CHECK (available_seats <= max_passengers),
    CONSTRAINT future_departure CHECK (departure_time > created_at)
);

-- Create ride requests table
CREATE TABLE IF NOT EXISTS ride_requests (
    request_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL REFERENCES rides(ride_id),
    rider_id UUID NOT NULL REFERENCES users(user_id),
    pickup_address TEXT NOT NULL,
    pickup_latitude DECIMAL(9,6) NOT NULL,
    pickup_longitude DECIMAL(9,6) NOT NULL,
    dropoff_address TEXT,
    dropoff_latitude DECIMAL(9,6),
    dropoff_longitude DECIMAL(9,6),
    status request_status DEFAULT 'pending',
    seats_requested INTEGER NOT NULL DEFAULT 1,
    distance_added_meters FLOAT,
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_seats_requested CHECK (seats_requested > 0)
);

-- Create the remaining tables
CREATE TABLE IF NOT EXISTS ride_passengers (
    ride_id UUID NOT NULL REFERENCES rides(ride_id),
    user_id UUID NOT NULL REFERENCES users(user_id),
    request_id UUID NOT NULL REFERENCES ride_requests(request_id),
    seats_taken INTEGER NOT NULL DEFAULT 1,
    pickup_time TIMESTAMP WITH TIME ZONE,
    dropoff_time TIMESTAMP WITH TIME ZONE,
    payment_status BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ride_id, user_id),
    CONSTRAINT valid_seats_taken CHECK (seats_taken > 0)
);

CREATE TABLE IF NOT EXISTS ratings (
    rating_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL REFERENCES rides(ride_id),
    rater_id UUID NOT NULL REFERENCES users(user_id),
    rated_id UUID NOT NULL REFERENCES users(user_id),
    rating DECIMAL(3,2) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_rating CHECK (rating >= 1 AND rating <= 5),
    CONSTRAINT different_users CHECK (rater_id != rated_id),
    UNIQUE(ride_id, rater_id, rated_id)
);

-- Function to calculate distance between two points using Haversine formula
CREATE OR REPLACE FUNCTION calculate_distance(
    lat1 FLOAT,
    lon1 FLOAT,
    lat2 FLOAT,
    lon2 FLOAT
) RETURNS FLOAT AS $$
DECLARE
    R FLOAT := 6371000; -- Earth radius in meters
    phi1 FLOAT;
    phi2 FLOAT;
    delta_phi FLOAT;
    delta_lambda FLOAT;
    a FLOAT;
    c FLOAT;
    d FLOAT;
BEGIN
    -- Convert latitude and longitude from degrees to radians
    phi1 := RADIANS(lat1);
    phi2 := RADIANS(lat2);
    delta_phi := RADIANS(lat2 - lat1);
    delta_lambda := RADIANS(lon2 - lon1);
    
    -- Haversine formula
    a := SIN(delta_phi/2) * SIN(delta_phi/2) +
         COS(phi1) * COS(phi2) *
         SIN(delta_lambda/2) * SIN(delta_lambda/2);
    c := 2 * ATAN2(SQRT(a), SQRT(1-a));
    d := R * c;
    
    RETURN d;
END;
$$ LANGUAGE plpgsql;

-- Function to find nearby rides
CREATE OR REPLACE FUNCTION find_nearby_rides(
    origin_lat FLOAT,
    origin_lon FLOAT,
    destination_lat FLOAT,
    destination_lon FLOAT,
    radius_meters FLOAT DEFAULT 5000,
    departure_after TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
) RETURNS TABLE (
    ride_id UUID,
    host_id UUID,
    origin_address TEXT,
    destination_address TEXT,
    departure_time TIMESTAMP WITH TIME ZONE,
    available_seats INTEGER,
    distance_from_origin FLOAT,
    distance_from_destination FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        r.ride_id,
        r.host_id,
        r.origin_address,
        r.destination_address,
        r.departure_time,
        r.available_seats,
        calculate_distance(origin_lat, origin_lon, r.origin_latitude, r.origin_longitude) AS distance_from_origin,
        calculate_distance(destination_lat, destination_lon, r.destination_latitude, r.destination_longitude) AS distance_from_destination
    FROM rides r
    WHERE r.status = 'scheduled'
      AND r.departure_time > departure_after
      AND r.available_seats > 0
      AND calculate_distance(origin_lat, origin_lon, r.origin_latitude, r.origin_longitude) <= radius_meters
      AND calculate_distance(destination_lat, destination_lon, r.destination_latitude, r.destination_longitude) <= radius_meters
    ORDER BY r.departure_time ASC;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS rides_departure_time_idx ON rides(departure_time);
CREATE INDEX IF NOT EXISTS rides_status_idx ON rides(status);
CREATE INDEX IF NOT EXISTS rides_host_id_idx ON rides(host_id);
CREATE INDEX IF NOT EXISTS ride_requests_ride_id_idx ON ride_requests(ride_id);
CREATE INDEX IF NOT EXISTS ride_requests_rider_id_idx ON ride_requests(rider_id);
CREATE INDEX IF NOT EXISTS ride_requests_status_idx ON ride_requests(status);
CREATE INDEX IF NOT EXISTS rides_origin_lat_long_idx ON rides(origin_latitude, origin_longitude);
CREATE INDEX IF NOT EXISTS rides_dest_lat_long_idx ON rides(destination_latitude, destination_longitude);

-- Triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicles_updated_at
BEFORE UPDATE ON vehicles
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rides_updated_at
BEFORE UPDATE ON rides
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ride_requests_updated_at
BEFORE UPDATE ON ride_requests
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Update available seats when a request is accepted
CREATE OR REPLACE FUNCTION update_available_seats()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'accepted' AND OLD.status = 'pending' THEN
        UPDATE rides
        SET available_seats = available_seats - NEW.seats_requested
        WHERE ride_id = NEW.ride_id;
    ELSIF NEW.status = 'rejected' AND OLD.status = 'accepted' THEN
        UPDATE rides
        SET available_seats = available_seats + NEW.seats_requested
        WHERE ride_id = NEW.ride_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_seats_on_request_status_change
AFTER UPDATE ON ride_requests
FOR EACH ROW
WHEN (OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION update_available_seats();

--
-- PostgreSQL database dump complete
--

