package openskyapi_test

import (
	"goosnapi/pkg/openskyapi"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRawState(t *testing.T) {
	rawStates := [][]interface{}{
		{
			"7c7f38",
			"ZEY     ",
			"Australia",
			1643623948,
			1643623948,
			144.7732,
			-37.9141,
			403.86,
			false,
			50.46,
			357.66,
			-0.65,
			nil,
			312.42,
			"2502",
			false,
			3,
		},
	}
	raw := openskyapi.NotStructuredOpenSkyResponse{
		Time:   1643623949,
		States: rawStates,
	}
	expected := openskyapi.StructuredOpenSkyResponse{
		Time: time.Unix(int64(1643623949), 0),
		States: []openskyapi.State{
			{
				ICAO24:             "7c7f38",
				CallSign:           "ZEY     ",
				OriginCountry:      "Australia",
				TimePosition:       1643623948,
				LastContact:        1643623948,
				Longitude:          144.7732,
				Latitude:           -37.9141,
				GeoAltitude:        403.86,
				OnGround:           false,
				Velocity:           50.46,
				Heading:            357.66,
				VerticalRate:       -0.65,
				BarometricAltitude: 312.42,
				Squawk:             "2502",
				Spi:                false,
				PositionSource:     "FLARM",
			},
		},
	}
	tovalidate := openskyapi.StructureRawResponse(raw)
	assert.Equal(t, expected, tovalidate)
}

func TestGetFlightsInArea(t *testing.T) {
	coverAreaTest := openskyapi.BoundBox{
		MinLatitude:  -40.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 139.147805,
		MaxLongitude: 154.590532,
	}
	nsrTest, nsrError := openskyapi.GetFlightsInArea(coverAreaTest)
	assert.Nil(t, nsrError)
	assert.NotNil(t, nsrTest)
}

func TestPrint(t *testing.T) {
	StructuredOpenSkyResponse := openskyapi.StructuredOpenSkyResponse{
		Time: time.Unix(int64(1643623949), 0),
		States: []openskyapi.State{
			{
				ICAO24:             "7c7f38",
				CallSign:           "ZEY     ",
				OriginCountry:      "Australia",
				TimePosition:       1643623948,
				LastContact:        1643623948,
				Longitude:          144.7732,
				Latitude:           -37.9141,
				GeoAltitude:        403.86,
				OnGround:           false,
				Velocity:           50.46,
				Heading:            357.66,
				VerticalRate:       -0.65,
				BarometricAltitude: 312.42,
				Squawk:             "2502",
				Spi:                false,
				PositionSource:     "FLARM",
			},
		},
	}
	err := StructuredOpenSkyResponse.Print()
	assert.Nil(t, err)
}

func TestBoundBoxValidate(t *testing.T) {
	coverAreaTestValid := openskyapi.BoundBox{
		MinLatitude:  -40.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 139.147805,
		MaxLongitude: 154.590532,
	}
	coverAreaTestInvalidRangeLatitude := openskyapi.BoundBox{
		MinLatitude:  200.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 139.147805,
		MaxLongitude: 154.590532,
	}
	coverAreaTestInvalidRangeLongitude := openskyapi.BoundBox{
		MinLatitude:  -40.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 200.147805,
		MaxLongitude: 154.590532,
	}
	coverAreaTestInvalidLatitude := openskyapi.BoundBox{
		MinLatitude:  -20.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 139.147805,
		MaxLongitude: 154.590532,
	}
	coverAreaTestInvalidLongitude := openskyapi.BoundBox{
		MinLatitude:  -40.110403,
		MaxLatitude:  -24.267845,
		MinLongitude: 160.147805,
		MaxLongitude: 154.590532,
	}
	err := coverAreaTestValid.Validate()
	assert.Nil(t, err)
	err = coverAreaTestInvalidRangeLatitude.Validate()
	assert.NotNil(t, err)
	err = coverAreaTestInvalidRangeLongitude.Validate()
	assert.NotNil(t, err)
	err = coverAreaTestInvalidLatitude.Validate()
	assert.NotNil(t, err)
	err = coverAreaTestInvalidLongitude.Validate()
	assert.NotNil(t, err)
}
