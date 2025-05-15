package service

import (
	"context"
	//"encoding/json"
	"errors"
	"time"

	"search-service/internal/cache"
	pb "search-service/proto"
)

type SearchService struct {
	GeoClient   GeoDistanceClient
	RedisClient *cache.RedisClient
}

type GeoDistanceClient interface {
	FilterRides(ctx context.Context, point *pb.Point, rides []*pb.Ride, matchType pb.MatchType) ([]*pb.Ride, error)
}

type SearchInput struct {
	Start *pb.Point
	End   *pb.Point
}

func NewSearchService(geoClient GeoDistanceClient, redisClient *cache.RedisClient) *SearchService {
	return &SearchService{
		GeoClient:   geoClient,
		RedisClient: redisClient,
	}
}

func (s *SearchService) SearchRides(ctx context.Context, input SearchInput, allRides []*pb.Ride) ([]*pb.Ride, error) {
	if input.Start == nil || input.End == nil {
		return nil, errors.New("start and end points are required")
	}

	// Generate cache key
	startHash := cache.GenerateGeohash(input.Start.Lat, input.Start.Lng)
	endHash := cache.GenerateGeohash(input.End.Lat, input.End.Lng)
	cacheKey := cache.GenerateCacheKey(startHash, endHash)

	// Check cache first
	if cached, err := s.RedisClient.GetCachedSearch(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Start filtering using geo-distance-service
	startFiltered, err := s.GeoClient.FilterRides(ctx, input.Start, allRides, pb.MatchType_START)
	if err != nil {
		return nil, err
	}

	endFiltered, err := s.GeoClient.FilterRides(ctx, input.End, startFiltered, pb.MatchType_END)
	if err != nil {
		return nil, err
	}

	// Cache the result
	err = s.RedisClient.CacheSearchResult(ctx, cacheKey, endFiltered, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	return endFiltered, nil
}
