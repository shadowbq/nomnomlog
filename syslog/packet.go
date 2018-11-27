package syslog

import (
	"fmt"
	"strings"
	"time"
)

// A Packet represents an RFC5425 syslog message
type Packet struct {
	Severity Severity
	Facility Facility
	Hostname string
	Tag      string
	Time     time.Time
	Message  string
}

// like time.RFC3339Nano but with a limit of 6 digits in the SECFRAC part
const rfc5424time = "2006-01-02T15:04:05.999999Z07:00"

// The combined Facility and Severity of this packet. See RFC5424 for details.
func (p Packet) Priority() Priority {
	pp := ((int(p.Facility) * 8) + int(p.Severity))
	return Priority(pp)
}

func (p Packet) cleanMessage() string {
	s := strings.Replace(p.Message, "\n", " ", -1)
	s = strings.Replace(s, "\r", " ", -1)
	return strings.Replace(s, "\x00", " ", -1)
}

// Generate creates a RFC5424 syslog format string for this packet.
func (p Packet) Generate(max_size int) string {
	ts := p.Time.Format(rfc5424time)
	if max_size == 0 {
		return fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.cleanMessage())
	} else {
		msg := fmt.Sprintf("<%d>1 %s %s %s - - - %s", p.Priority(), ts, p.Hostname, p.Tag, p.cleanMessage())
		if len(msg) > max_size {
			return msg[0:max_size]
		} else {
			return msg
		}
	}
}

// A convenience function for testing (Syslog.Parse)
func Parse(line string) (Packet, error) {
	var (
		packet   Packet
		priority int
		ts       string
		hostname string
		tag      string
	)

	splitLine := strings.Split(line, " - - - ")
	if len(splitLine) != 2 {
		return packet, fmt.Errorf("couldn't parse syslog line: %s", line)
	}

	fmt.Sscanf(splitLine[0], "<%d>1 %s %s %s", &priority, &ts, &hostname, &tag)

	t, err := time.Parse(rfc5424time, ts)
	if err != nil {
		return packet, err
	}

	// bitwise operators
	// &    bitwise AND
	// >>   right shift
	//32 >> 5 is "32 divided by 2, 5 times"

	// Severity: SeverityMap(priority & 7),
	// Facility: FacilityMap(priority >> 3),

	return Packet{
		Severity: PriorityExtractSeverity(priority),
		Facility: PriorityExtractFacility(priority),
		Hostname: hostname,
		Tag:      tag,
		Time:     t,
		Message:  splitLine[1],
	}, nil
}
