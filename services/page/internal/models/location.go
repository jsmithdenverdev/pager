package models

type LocationType string

const (
	// LocationTypeAddress represents a street address location.
	//
	// E.g. 13001 Hwy 3, Evergreen, CO 80439.
	LocationTypeAddress LocationType = "LOC_ADDRESS"
	// LocationTypeCommonName represents a common name location.
	//
	// E.g. Echo Lake / Mt. Blue Sky.
	LocationTypeCommonName LocationType = "LOC_COMMON_NAME"
	// LocationTypeDD represents a set of coordinates in decimal degrees format.
	//
	// E.g. 39.660783°, -105.604551°.
	LocationTypeDD LocationType = "LOC_DECIMAL_DEGREES"
	// LocationTypeDDM represents a set of coordinates in degrees decimal minutes
	// format.
	//
	// E.g. 39° 39.647' N, 105° 36.2731' W.
	LocationTypeDDM LocationType = "LOC_DEGREES_DECIMAL_MINUTES"
	// LocationTypeDMS represents a set of coordinates in degrees minutes seconds
	// format.
	//
	// E.g. 39° 39' 38.82" N, 105° 36' 16.386" W
	LocationTypeDMS LocationType = "LOC_DEGREES_MINUTES_SECONDS"
)

// A Location is container for location metadata related to a Page.
type Location struct {
	Type LocationType
	Data string
}
