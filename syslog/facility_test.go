package syslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupFacility(t *testing.T) {
	assert := assert.New(t)

	var facility Facility
	var err error

	facility, err = FacilityMap("local1")
	if facility != LogLocal1 && err != nil {
		t.Errorf("Failed to lookup facility local1")
	}

	facility, err = FacilityMap("foo")
	if facility != 0 && err != ErrFacility {
		t.Errorf("Failed to lookup facility foo")
	}

	facility, err = FacilityMap("")
	if facility != 0 && err != ErrFacility {
		t.Errorf("Failed to lookup empty facility")
	}

	facility_str, ok := Facilitykeymap(facilities, 11)
	assert.Equal(facility_str, "ftp")
	assert.Equal(ok, true)
}
