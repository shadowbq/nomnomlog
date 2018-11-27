package syslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFacility(t *testing.T) {

	//Priority(68) =  Severity(Warning|4) and Facility(UUCP|8)
	var facility Facility

	facility = PriorityExtractFacility(68)
	assert.Equal(t, int(facility), 8)

}

func TestExtractSeverity(t *testing.T) {

	//Priority(68) =  Severity(Warning|4) and Facility(UUCP|8)
	var severity Severity

	severity = PriorityExtractSeverity(68)
	assert.Equal(t, int(severity), 4)

}
