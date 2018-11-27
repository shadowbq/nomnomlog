package syslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupSeverity(t *testing.T) {
	assert := assert.New(t)

	var sev Severity
	var err error

	sev, err = SeverityMap("warn")
	if sev != SevWarning && err != nil {
		t.Errorf("Failed to lookup severity warning")
	}

	sev, err = SeverityMap("foo")
	if sev != 0 && err != ErrSeverity {
		t.Errorf("Failed to lookup severity foo")
	}

	sev, err = SeverityMap("")
	if sev != 0 && err != ErrSeverity {
		t.Errorf("Failed to lookup empty severity")
	}

	severity_str, ok := Severitykeymap(severities, 5)
	assert.Equal(severity_str, "notice")
	assert.Equal(ok, true)

}
