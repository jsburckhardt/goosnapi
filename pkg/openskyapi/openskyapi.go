package openskyapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Struct for capturing the raw response from OpenSky (raw)
type NotStructuredOpenSkyResponse struct {
	Time   int             `json:"time"`
	States [][]interface{} `json:"states"`
}

// Structured OpenSky Response
type StructuredOpenSkyResponse struct {
	Time   time.Time `json:"time"`
	States []State   `json:"states"`
}

// Structured OpenSky Response States
type State struct {
	ICAO24        string  `json:"icao24"`         // ICAO24 address of the transmitter in hex string representation.
	CallSign      string  `json:"callsign"`       // CallSign of the vehicle. Can be nil if no callsign has been received.
	OriginCountry string  `json:"origin_country"` // Inferred through the ICAO24 address.
	TimePosition  int     `json:"time_position"`  // UnixTime of last position report. Can be nil if there was no position report received by OpenSky within 15s before.
	LastContact   int     `json:"last_contact"`   // UnixTime of last received message from this transponder.
	Longitude     float64 `json:"longitude"`      // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	Latitude      float64 `json:"latitude"`       // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	GeoAltitude   float64 `json:"geo_altitude"`   // Geometric altitude in meters. Can be nil.
	OnGround      bool    `json:"on_ground"`      // True if aircraft is on ground (sends ADS-B surface position reports).
	Velocity      float64 `json:"velocity"`       // Velocity over ground in m/s. Can be nil if information not present.
	Heading       float64 `json:"heading"`        // Heading in decimal degrees (0 is north). Can be nil if information not present.
	VerticalRate  float64 `json:"vertical_rate"`  // In m/s, incline is positive, decline negative. Can be nil if information not present.
	// Sensors            []int   `json:"sensors"`         // Serial numbers of sensors which received messages from the vehicle within the validity period of this state vector. Can be nil if no filtering for sensor has been requested.
	BarometricAltitude float64 `json:"baro_altitude"`   // Barometric altitude in meters. Can be nil.
	Squawk             string  `json:"squawk"`          // Transponder code aka Squawk. Can be empty.
	Spi                bool    `json:"spi"`             // Special purpose indicator.
	PositionSource     string  `json:"position_source"` // Origin of this state’s position.
}

// BoundBox is the area used for monitoring flights (states)
type BoundBox struct {
	MinLatitude  float64 `mapstructure:"minlatitude"`
	MaxLatitude  float64 `mapstructure:"maxlatitude"`
	MinLongitude float64 `mapstructure:"minlongitude"`
	MaxLongitude float64 `mapstructure:"maxlongitude"`
}

// PositionSource is a map for identifying the differnt type of sources in the flights
var PositionSource = map[int]string{
	0: "ADSB",
	1: "ASTERIX",
	2: "MLAT",
	3: "FLARM",
}

// GetFlightsInArea is the function incharge of consuming OpensSky API and returning the flights in the area. The function returns NonStructuredOpenSkyResponse (raw response)
func GetFlightsInArea(coverArea BoundBox) (nsr NotStructuredOpenSkyResponse, err error) {
	url := fmt.Sprintf("https://opensky-network.org/api/states/all?lamin=%s&lomin=%s&lamax=%s&lomax=%s", fmt.Sprintf("%.4f", coverArea.MinLatitude), fmt.Sprintf("%.4f", coverArea.MinLongitude), fmt.Sprintf("%.4f", coverArea.MaxLatitude), fmt.Sprintf("%.4f", coverArea.MaxLongitude))
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error making the request. error: %+v\n", err)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &nsr)
	return
}

// StructureRawResponse is the function incharge of structuring the raw response from the OpensSky API
func StructureRawResponse(nsr NotStructuredOpenSkyResponse) (sr StructuredOpenSkyResponse) {
	sr.Time = time.Unix(int64(nsr.Time), 0)
	for _, state := range nsr.States {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		ss := State{
			ICAO24:        validateString(state[0]),
			CallSign:      validateString(state[1]),
			OriginCountry: validateString(state[2]),
			TimePosition:  int(validateFloat64(state[3])), //nullable
			LastContact:   int(validateFloat64(state[4])),
			Longitude:     validateFloat64(state[5]), //nullable
			Latitude:      validateFloat64(state[6]), //nullable
			GeoAltitude:   validateFloat64(state[7]), //nullable
			OnGround:      validateBool(state[8]),
			Velocity:      validateFloat64(state[9]),  //nullable
			Heading:       validateFloat64(state[10]), //nullable
			VerticalRate:  validateFloat64(state[11]), //nullable
			// Sensors:            make([]int, len(state[12].([]interface{}))),
			BarometricAltitude: validateFloat64(state[13]), //nullable
			Squawk:             validateString(state[14]),
			Spi:                validateBool(state[15]),
			PositionSource:     PositionSource[int(validateFloat64(state[16]))],
		}
		sr.States = append(sr.States, ss)
	}
	return
}

// validateString internal function to validate the empty interface to string
func validateString(i interface{}) (value string) {
	switch f := i.(type) {
	case string:
		value = f
	case int:
		value = strconv.Itoa(f)
	default:
		value = ""
	}
	return
}

// validateFloat64 internal function to validate the empty interface to float64
func validateFloat64(f interface{}) (value float64) {
	switch f := f.(type) {
	case float64:
		value = f
	case int:
		value = float64(f)
	}
	return
}

// validateBool internal function to validate the empty interface to bool
func validateBool(f interface{}) (value bool) {
	switch f := f.(type) {
	case bool:
		value = f
	}
	return
}

// Print simple method to print the StructuredOpenSkyResponse
func (f StructuredOpenSkyResponse) Print() (err error) {
	fmt.Printf("Got %d records.\n", len(f.States))
	fmt.Printf("Time: %v\n", f.Time)
	template := `ICAO24: %v	 CallSign: %v	OriginCountry: %v	TimePosition: %v	LastContact: %v	Longitude: %v	Latitude: %v	GeoAltitude: %v	OnGround: %v	Velocity: %v	Heading: %v	VerticalRate: %v	BarometricAltitude: %v	Squawk: %v	Spi: %v	PositionSource: %v`
	for _, s := range f.States {
		_, err = fmt.Println(fmt.Sprintf(template, s.ICAO24, s.CallSign, s.OriginCountry, s.TimePosition, s.LastContact, s.Longitude, s.Latitude, s.GeoAltitude, s.OnGround, s.Velocity, s.Heading, s.VerticalRate, s.BarometricAltitude, s.Squawk, s.Spi, s.PositionSource))
		if err != nil {
			return
		}
	}
	return
}

// Validate limits of the box
func (b BoundBox) Validate() (err error) {
	if b.MinLongitude >= b.MaxLongitude {
		err = fmt.Errorf("MinLongitude is greater than MaxLongitude. MinLongitude: %v, MaxLongitude: %v", b.MinLongitude, b.MaxLongitude)
		return
	}
	if b.MinLatitude > b.MaxLatitude {
		err = fmt.Errorf("MinLatitude is greater than MaxLatitude. MinLatitude: %v, MaxLatitude: %v", b.MinLatitude, b.MaxLatitude)
		return
	}
	if b.MinLongitude > 180 || b.MaxLongitude > 180 || b.MinLongitude < -180 || b.MaxLongitude < -180 {
		err = fmt.Errorf("Longitude is out of range (-180, 180). MinLongitude: %v, MaxLongitude: %v", b.MinLongitude, b.MaxLongitude)
		return
	}
	if b.MinLatitude > 90 || b.MaxLatitude > 90 || b.MinLatitude < -90 || b.MaxLatitude < -90 {
		err = fmt.Errorf("Latitude is out of range (-90, 90). MinLatitude: %v, MaxLatitude: %v", b.MinLatitude, b.MaxLatitude)
		return
	}
	return
}
