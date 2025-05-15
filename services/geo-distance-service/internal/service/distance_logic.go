package service

import (
	"geo-distance-service/internal/config"
	"geo-distance-service/proto"
	"math"
	"strings"
)

const EarthRadiusKm = 6371

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusKm * c
}

func FilterRidesByProximity(cfg *config.Config, input *proto.FilterRequest) []*proto.Ride {
	var filtered []*proto.Ride
	point := input.Point
	matchType := strings.ToLower(input.MatchType.String())

	for _, ride := range input.Rides {
		var rLat, rLng float64
		if matchType == "start" {
			rLat, rLng = ride.StartPoint.Lat, ride.StartPoint.Lng
		} else {
			rLat, rLng = ride.EndPoint.Lat, ride.EndPoint.Lng
		}
		d := Haversine(point.Lat, point.Lng, rLat, rLng)
		if d <= cfg.RadiusKM {
			filtered = append(filtered, ride)
		}
	}
	return filtered
}
