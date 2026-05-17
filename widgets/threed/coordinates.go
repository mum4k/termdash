// threed/coordinate.go

package threed

import (
	"math"
	"strconv"
	"strings"
)

// Coordinate represents a single geographic coordinate.
type Coordinate struct {
	Longitude float64 // Longitude in degrees
	Latitude  float64 // Latitude in degrees
	Altitude  float64 // Altitude in meters
}

// ParseCoordinates parses a KML coordinates string into a slice of Coordinates.
func ParseCoordinates(coordStr string) ([]Coordinate, error) {
	coordStr = strings.TrimSpace(coordStr)
	coordPairs := strings.Fields(coordStr)
	coordinates := make([]Coordinate, 0, len(coordPairs))

	for _, pair := range coordPairs {
		values := strings.Split(pair, ",")
		if len(values) < 2 {
			continue
		}

		longitude, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return nil, err
		}

		latitude, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			return nil, err
		}

		altitude := 0.0
		if len(values) > 2 {
			altitude, err = strconv.ParseFloat(values[2], 64)
			if err != nil {
				return nil, err
			}
		}

		coordinates = append(coordinates, Coordinate{
			Longitude: longitude,
			Latitude:  latitude,
			Altitude:  altitude,
		})
	}

	return coordinates, nil
}

const (
	earthRadius = 6371.0 // Earth's radius in kilometers
	scaleFactor = 0.0001 // Scale factor to adjust the size of the rendered model
)

// GeoTo3D converts geographical coordinates to 3D Cartesian coordinates.
func GeoTo3D(coord Coordinate) Vector3D {
	// Convert degrees to radians
	latRad := coord.Latitude * math.Pi / 180.0
	lonRad := coord.Longitude * math.Pi / 180.0

	// Simple equirectangular projection
	x := earthRadius * lonRad * math.Cos(latRad)
	y := earthRadius * latRad
	z := coord.Altitude / 1000.0 // Convert altitude from meters to kilometers

	// Apply scaling
	x *= scaleFactor
	y *= scaleFactor
	z *= scaleFactor

	return Vector3D{X: x, Y: y, Z: z}
}
