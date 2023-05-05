package incursions

import "math"

type SecurityClass string

const (
	HighSec SecurityClass = "High"
	LowSec  SecurityClass = "Low"
	NullSec SecurityClass = "Null"
)

func ParseSecurityClass(status float64) SecurityClass {
	roundedSecStatus := ccp_round(status)

	if roundedSecStatus >= .5 {
		return HighSec
	} else if roundedSecStatus >= .1 {
		return LowSec
	}
	return NullSec
}

func ccp_round(status float64) float64 {
	if status > 0.0 && status < 0.05 {
		return math.Ceil(status*10) / 10
	}

	return math.Round(status*10) / 10
}
