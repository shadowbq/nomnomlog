package syslog

import "fmt"

type Severity int

// Returned when looking up a non-existent facility or severity
var ErrSeverity = fmt.Errorf("Not a designated RFC5424 Severity")

// RFC5424 Severities
// iota represents successive untyped integer constants.
const (
	SevEmerg Severity = iota
	SevAlert
	SevCrit
	SevErr
	SevWarning
	SevNotice
	SevInfo
	SevDebug
)

// Severity Mapping 0 - 7
var severities = map[string]Severity{
	"emerg":  SevEmerg,
	"alert":  SevAlert,
	"crit":   SevCrit,
	"err":    SevErr,
	"warn":   SevWarning,
	"notice": SevNotice,
	"info":   SevInfo,
	"debug":  SevDebug,
}

// SeverityMap returns the int of the named severity as Severity type. It returns ErrSeverity if the severity
// does not exist.
func SeverityMap(name string) (Severity, error) {
	p, ok := severities[name]
	if !ok {
		return 0, ErrSeverity
	}
	return p, nil
}

// Given the Hashmap severities and an Severity int it returns string of the Severity type
func Severitykeymap(m map[string]Severity, value int) (key string, ok bool) {
	for k, v := range m {
		if int(v) == value {
			key = k
			ok = true
			return
		}
	}
	return
}
