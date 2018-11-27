package syslog

// A Syslog Priority is a combination of Severity and Facility normally 0 to 191

// RFC5424 - 6.2.1 "The number [.sic.] is known as the Priority value (PRIVAL) and represents both the
// Facility and Severity.  The Priority value consists of one, two, or three decimal integers. [...] The Priority
// value is calculated by first multiplying the Facility number by 8 and then adding the numerical value of the Severity"

// For example, a Priority value of 13 is “user”[1] Facility and “notice”[5] Severity. (1*8)+5=13

type Priority int

// PriorityExtractSeverity returns the key severity from the priority.
func PriorityExtractSeverity(priority int) Severity {
	sv := priority & 7
	return Severity(sv)
}

// PriorityExtractSeverity returns the key severity from the priority.
func PriorityExtractFacility(priority int) Facility {
	fv := priority >> 3
	return Facility(fv)
}
