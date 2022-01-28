package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	// "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type BoundBox struct {
	MinLatitude  float32 `mapstructure:"minlatitude"`
	MaxLatitude  float32 `mapstructure:"maxlatitude"`
	MinLongitude float32 `mapstructure:"minlongitude"`
	MaxLongitude float32 `mapstructure:"maxlongitude"`
}
type ConfigStructure struct {
	OpenSky   BoundBox `yaml:"opensky"`
	Frequency int      `mapstructure:"frequency"`
}

// Not Structure Open Sky Response
type NSOpenSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}
type OpenSkyResponse struct {
	Time   int64   `json:"time"`
	States []State `json:"states"`
}

var CONFIG = "config.yaml"

type PositionSource int

// All pointer fields are nullable, therefore checks are required, before accessing those fields.
type State struct {
	ICAO24             string         `json:"icao24"`                  // ICAO24 address of the transmitter in hex string representation.
	CallSign           string         `json:"callsign,omitempty"`      // CallSign of the vehicle. Can be nil if no callsign has been received.
	OriginCountry      string         `json:"origin_country"`          // Inferred through the ICAO24 address.
	TimePosition       *time.Time     `json:"time_position,omitempty"` // UnixTime of last position report. Can be nil if there was no position report received by OpenSky within 15s before.
	LastContact        time.Time      `json:"last_contact"`            // UnixTime of last received message from this transponder.
	Longitude          *float64       `json:"longitude,omitempty"`     // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	Latitude           *float64       `json:"latitude,omitempty"`      // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	GeoAltitude        *float64       `json:"geo_altitude,omitempty"`  // Geometric altitude in meters. Can be nil.
	OnGround           bool           `json:"on_ground"`               // True if aircraft is on ground (sends ADS-B surface position reports).
	Velocity           *float64       `json:"velocity,omitempty"`      // Velocity over ground in m/s. Can be nil if information not present.
	Heading            *float64       `json:"heading,omitempty"`       // Heading in decimal degrees (0 is north). Can be nil if information not present.
	VerticalRate       *float64       `json:"vertical_rate,omitempty"` // In m/s, incline is positive, decline negative. Can be nil if information not present.
	Sensors            []int          `json:"sensors,omitempty"`       // Serial numbers of sensors which received messages from the vehicle within the validity period of this state vector. Can be nil if no filtering for sensor has been requested.
	BarometricAltitude *float64       `json:"baro_altitude,omitempty"` // Barometric altitude in meters. Can be nil.
	Squawk             string         `json:"squawk,omitempty"`        // Transponder code aka Squawk. Can be empty.
	Spi                bool           `json:"spi"`                     // Special purpose indicator.
	PositionSource     PositionSource `json:"position_source"`         // Origin of this state’s position.
}

func main() {
	// load config
	viper.SetConfigType("yaml")
	viper.SetConfigFile(CONFIG)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	var t ConfigStructure
	err = viper.Unmarshal(&t)
	if err != nil {
		fmt.Println(err)
		return
	}

	// retrive data every "Frequency" seconds
	var sr OpenSkyResponse
	var ur NSOpenSkyResponse
	for {
		url := fmt.Sprintf("https://opensky-network.org/api/states/all?lamin=%s&lomin=%s&lamax=%s&lomax=%s", fmt.Sprintf("%.4f", t.OpenSky.MinLatitude), fmt.Sprintf("%.4f", t.OpenSky.MinLongitude), fmt.Sprintf("%.4f", t.OpenSky.MaxLatitude), fmt.Sprintf("%.4f", t.OpenSky.MaxLongitude))
		req, _ := http.NewRequest("GET", url, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(body, &ur)
		if err != nil {
			fmt.Println(err)
		}

		// Process the response
		parseNonStructuredResponseToStructureResponse(&ur, &sr)
		printStructuredResponse(sr)

		// wait t.frequency seconds
		time.Sleep(time.Duration(t.Frequency) * time.Second)
	}
}

func printStructuredResponse(sr OpenSkyResponse) {
	fmt.Printf("Got %d records.\n", len(sr.States))
	fmt.Printf("Time: ", sr.Time)
	for _, s := range sr.States {
		fmt.Printf("ICAO24: %+v\t", s.ICAO24)
		fmt.Printf("CallSign: %+v\t", s.CallSign)
		fmt.Printf("OriginCountry: %+v\t", s.OriginCountry)
		fmt.Printf("TimePosition: %+v\t", (*s.TimePosition).String())
		fmt.Printf("LastContact: %+v\t", s.LastContact)
		fmt.Printf("Longitude: %+v\t", *s.Longitude)
		fmt.Printf("Latitude: %+v\t", *s.Latitude)
		fmt.Printf("GeoAltitude: %+v\t", *s.GeoAltitude)
		fmt.Printf("OnGround: %+v\t", s.OnGround)
		fmt.Printf("Velocity: %+v\t", *s.Velocity)
		fmt.Printf("Heading: %+v\t", *s.Heading)
		fmt.Printf("VerticalRate: %+v\t", *s.VerticalRate)
		fmt.Printf("Sensors: %+v\t", s.Sensors)
		fmt.Printf("BarometricAltitude: %+v\t", *s.BarometricAltitude)
		fmt.Printf("Squawk: %+v\t", s.Squawk)
		fmt.Printf("Spi: %+v\t", s.Spi)
		fmt.Printf("PositionSource: %+v\n", s.PositionSource)
	}
}

func parseNonStructuredResponseToStructureResponse(nonStructuredOpenSkyResponse *NSOpenSkyResponse, structuredOpenSkyResponse *OpenSkyResponse) {
	structuredOpenSkyResponse.Time = nonStructuredOpenSkyResponse.Time
	for i, unstructureState := range nonStructuredOpenSkyResponse.States {
		var structuredState State
		structuredState, err := parseState(unstructureState, i)
		if err != nil {
			fmt.Println(err)
		}
		structuredOpenSkyResponse.States = append(structuredOpenSkyResponse.States, structuredState)
	}
}

// Parse a single state array from an unstructured states response.
// The i parameter represents the index of the state element in the states response.
func parseState(s []interface{}, i int) (state State, err error) {
	if len(s) < 17 {
		err = fmt.Errorf("invalid state object at position %v: response contains %v values, expected 17", i, len(s))
		return
	}
	// icao24
	icao24, ok := s[0].(string)
	if !ok {
		err = fmt.Errorf("invalid icao24 value at position %d: %v", i, s[0])
		return
	}
	// callsign
	var callsign string
	if s[1] != nil {
		callsign, ok = s[1].(string)
		if !ok {
			err = fmt.Errorf("invalid callsign value at position %d: %v", i, s[1])
			return
		}
	}
	// origin_country
	originCountry, ok := s[2].(string)
	if !ok {
		err = fmt.Errorf("invalid origin_country value at position %d: %v", i, s[2])
		return
	}
	// time_position
	var rawTimePosition int64
	var timePosition *time.Time
	if s[3] != nil {
		rawTimePosition, err = jsonNumberToInt(s[3])
		if err != nil {
			err = fmt.Errorf("invalid time_position value at position %d: %w", i, err)
			return
		}
		unixTime := time.Unix(rawTimePosition, 0)
		timePosition = &unixTime
	}
	// last_contact
	var lastContact int64
	lastContact, err = jsonNumberToInt(s[4])
	if err != nil {
		err = fmt.Errorf("invalid last_contact value at position %d: %w", i, err)
		return
	}
	// longitude
	var lon *float64
	if rawLon, ok := s[5].(float64); ok {
		lon = &rawLon
	}
	// latitude
	var lat *float64
	if rawLat, ok := s[6].(float64); ok {
		lat = &rawLat
	}
	// baro_altitude
	var baroAltitude *float64
	if rawBaroAltitude, ok := s[7].(float64); ok {
		baroAltitude = &rawBaroAltitude
	} else {
		rawBaroAltitude = float64(0)
		baroAltitude = &rawBaroAltitude
	}
	// on_ground
	onGround, ok := s[8].(bool)
	if !ok {
		err = fmt.Errorf("invalid on_ground value at position %d: %v", i, s[8])
		return
	}
	// velocity
	var velocity *float64
	if rawVelocity, ok := s[9].(float64); ok {
		velocity = &rawVelocity
	}
	// true_track
	var trueTrack *float64
	if rawTrueTrack, ok := s[10].(float64); ok {
		trueTrack = &rawTrueTrack
	}
	// vertical_rate
	var verticalRate *float64
	if rawVerticalRate, ok := s[11].(float64); ok {
		verticalRate = &rawVerticalRate
	} else {
		rawVerticalRate = float64(0)
		verticalRate = &rawVerticalRate
	}
	// sensors
	var sensors []int
	if s[12] != nil {
		sensors, err = jsonNumberArrayToIntArray(s[12])
		if err != nil {
			err = fmt.Errorf("invalid sensors value at position %d: %w", i, err)
			return
		}
	}
	// geo_altitude
	var geoAltitude *float64
	if rawGeoAltitude, ok := s[13].(float64); ok {
		geoAltitude = &rawGeoAltitude
	} else {
		rawGeoAltitude = float64(0)
		geoAltitude = &rawGeoAltitude
	}
	// squawk
	var squawk string
	if s[14] != nil {
		squawk, ok = s[14].(string)
		if !ok {
			err = fmt.Errorf("invalid squawk value at position %d: %v", i, s[14])
			return
		}
	}
	// spi
	spi, ok := s[15].(bool)
	if !ok {
		err = fmt.Errorf("invalid spi value at position %d: %v", i, s[15])
		return
	}
	// position_source
	var positionSource int64
	positionSource, err = jsonNumberToInt(s[16])
	if err != nil {
		err = fmt.Errorf("invalid position_source value at position %d: %w", i, err)
		return
	}
	// Set state values
	state = State{
		ICAO24:             icao24,
		CallSign:           callsign,
		OriginCountry:      originCountry,
		TimePosition:       timePosition,
		LastContact:        time.Unix(lastContact, 0),
		Longitude:          lon,
		Latitude:           lat,
		GeoAltitude:        geoAltitude,
		OnGround:           onGround,
		Velocity:           velocity,
		Heading:            trueTrack,
		VerticalRate:       verticalRate,
		Sensors:            sensors,
		BarometricAltitude: baroAltitude,
		Squawk:             squawk,
		Spi:                spi,
		PositionSource:     PositionSource(positionSource),
	}
	return
}

// Helper function to convert a number received in a json object to an int64 type.
// Throws an error, if the number could not be parsed.
func jsonNumberToInt(val interface{}) (i int64, err error) {
	fVal, ok := val.(float64)
	if !ok {
		err = fmt.Errorf("couldn't parse %v as number", val)
		return
	}
	i = int64(fVal)
	return
}

// Helper function to convert a number array received in a json object to an []int type.
// Throws an error, if the value could not be parsed as a number array.
func jsonNumberArrayToIntArray(val interface{}) (a []int, err error) {
	aVal, ok := val.([]float64)
	if !ok {
		err = fmt.Errorf("couldn't parse %v as number array", val)
		return
	}
	for _, v := range aVal {
		a = append(a, int(v))
	}
	return
}
