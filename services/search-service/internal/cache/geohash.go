package cache

import (
	"github.com/mmcloughlin/geohash"
)

func GenerateGeohash(lat, lng float64) string {
	// Use precision 6 for ~1.2km accuracy
	return geohash.EncodeWithPrecision(lat, lng, 6)
}
