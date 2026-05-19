// threed/kmlparser.go

package threed

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
)

// KML represents the root element of a KML file.
type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

// Document represents a KML Document element.
type Document struct {
	XMLName    xml.Name    `xml:"Document"`
	Name       string      `xml:"name"`
	Placemarks []Placemark `xml:"Placemark"`
	Folders    []Folder    `xml:"Folder"`
}

// Folder represents a KML Folder element.
type Folder struct {
	XMLName    xml.Name    `xml:"Folder"`
	Name       string      `xml:"name"`
	Placemarks []Placemark `xml:"Placemark"`
	Folders    []Folder    `xml:"Folder"`
}

// Placemark represents a KML Placemark element.
type Placemark struct {
	XMLName     xml.Name    `xml:"Placemark"`
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Point       *Point      `xml:"Point"`
	LineString  *LineString `xml:"LineString"`
	Polygon     *Polygon    `xml:"Polygon"`
}

// Point represents a KML Point element.
type Point struct {
	XMLName     xml.Name `xml:"Point"`
	Coordinates string   `xml:"coordinates"`
}

// LineString represents a KML LineString element.
type LineString struct {
	XMLName     xml.Name `xml:"LineString"`
	Coordinates string   `xml:"coordinates"`
}

// Polygon represents a KML Polygon element.
type Polygon struct {
	XMLName       xml.Name      `xml:"Polygon"`
	OuterBoundary OuterBoundary `xml:"outerBoundaryIs"`
}

// OuterBoundary represents the outer boundary of a KML Polygon.
type OuterBoundary struct {
	XMLName    xml.Name   `xml:"outerBoundaryIs"`
	LinearRing LinearRing `xml:"LinearRing"`
}

// LinearRing represents a KML LinearRing element.
type LinearRing struct {
	XMLName     xml.Name `xml:"LinearRing"`
	Coordinates string   `xml:"coordinates"`
}

// FetchAndParseKML fetches the KML file from the given URL and parses it.
func FetchAndParseKML(ctx context.Context, url string, logger *log.Logger) (*KML, error) {
	logger.Printf("Fetching KML from URL: %s", url)

	// Fetch the KML file.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Printf("Failed to create HTTP request: %v", err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("Failed to fetch KML file: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Failed to read KML data: %v", err)
		return nil, err
	}
	logger.Printf("Successfully fetched KML data, size: %d bytes", len(data))

	// Parse the KML file.
	var kml KML
	if err := xml.Unmarshal(data, &kml); err != nil {
		logger.Printf("Failed to parse KML data: %v", err)
		return nil, err
	}

	logger.Printf("Successfully parsed KML data")

	return &kml, nil
}

// ExtractPlacemarksFromDocument extracts placemarks from a KML document.
func ExtractPlacemarksFromDocument(doc Document) []Placemark {
	var placemarks []Placemark
	placemarks = append(placemarks, doc.Placemarks...)
	for _, folder := range doc.Folders {
		placemarks = append(placemarks, ExtractPlacemarksFromFolder(folder)...)
	}
	return placemarks
}

// ExtractPlacemarksFromFolder extracts placemarks from a KML folder.
func ExtractPlacemarksFromFolder(folder Folder) []Placemark {
	var placemarks []Placemark
	placemarks = append(placemarks, folder.Placemarks...)
	for _, subFolder := range folder.Folders {
		placemarks = append(placemarks, ExtractPlacemarksFromFolder(subFolder)...)
	}
	return placemarks
}

// GenerateModelFromKML generates a 3D model from KML data.
func GenerateModelFromKML(kml *KML, logger *log.Logger) (*Model, error) {
	if kml == nil {
		return nil, fmt.Errorf("kml is nil")
	}
	if logger == nil {
		logger = quietLogger()
	}
	model := NewModel()
	placemarks := ExtractPlacemarksFromDocument(kml.Document)

	logger.Printf("Found %d placemarks in KML", len(placemarks))

	for _, placemark := range placemarks {
		// Handle Point geometries
		if placemark.Point != nil && placemark.Point.Coordinates != "" {
			coords, err := ParseCoordinates(placemark.Point.Coordinates)
			if err != nil {
				logger.Printf("Failed to parse Point coordinates: %v", err)
				return nil, err
			}
			if len(coords) > 0 {
				position := GeoTo3D(coords[0])
				marker := createMarker(position, 0.1, '●') // Using a small size for markers
				model.Append(marker)
			}
		}

		// Handle LineString geometries
		if placemark.LineString != nil && placemark.LineString.Coordinates != "" {
			coordsList, err := ParseCoordinates(placemark.LineString.Coordinates)
			if err != nil {
				logger.Printf("Failed to parse LineString coordinates: %v", err)
				return nil, err
			}
			coords := make([]Vector3D, len(coordsList))
			for i, coord := range coordsList {
				coords[i] = GeoTo3D(coord)
			}
			line := createLine(coords, '─') // Using a line character
			model.Append(line)
		}

		// Handle Polygon geometries
		if placemark.Polygon != nil && placemark.Polygon.OuterBoundary.LinearRing.Coordinates != "" {
			coordsList, err := ParseCoordinates(placemark.Polygon.OuterBoundary.LinearRing.Coordinates)
			if err != nil {
				logger.Printf("Failed to parse Polygon coordinates: %v", err)
				return nil, err
			}
			coords := make([]Vector3D, len(coordsList))
			for i, coord := range coordsList {
				coords[i] = GeoTo3D(coord)
			}
			polygon := createPolygon(coords, '█') // Using a block character
			model.Append(polygon)
		}
	}
	center := model.Center()
	model.Translate(center)
	logger.Printf("Model centered at (%.2f, %.2f, %.2f)", center.X, center.Y, center.Z)

	logger.Printf("Generated model with %d faces", len(model.Faces))

	return model, nil
}
