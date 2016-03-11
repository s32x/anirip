package anirip

import (
	"strconv"
	"time"
)

// Shifts the subtitle time to account for the passed millisecond sub offset
func ShiftTime(subTime string, offset int) (string, error) {
	// Sets the parsing format to accept a time like this
	assFormat := "15:04:05.999999"

	// Parses the passed subtitle time
	tm, err := time.Parse(assFormat, subTime)
	if err != nil {
		return "", Error{Message: "There was an error parsing subtitle time", Err: err}
	}

	// Parses the offset to a duration that will be subtracted from the parsed sub time
	offsetDuration, err := time.ParseDuration("-" + strconv.Itoa(offset) + "ms")
	if err != nil {
		return "", Error{Message: "There was an error parsing subtitle time", Err: err}
	}
	tm = tm.Add(offsetDuration)

	// returns the new shifted time
	return tm.Format("15:04:05.00"), nil
}
