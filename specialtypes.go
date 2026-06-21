package t1y

import (
	"fmt"
	"regexp"

	json "github.com/goccy/go-json"
)

// objectIDPattern is the regex pattern for valid ObjectID hex strings.
var objectIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// ==================== ObjectID ====================

// ObjectID creates an ObjectID marker string that the server will convert to a MongoDB ObjectID.
// The id parameter must be a 24-character hexadecimal string.
func ObjectID(id string) string {
	if !objectIDPattern.MatchString(id) {
		panic(fmt.Sprintf("Invalid ObjectID: \"%s\" (must be 24 hex characters)", id))
	}
	return fmt.Sprintf("ObjectID('%s')", id)
}

// ==================== Date Types ====================

// Date creates a Date marker string. The server converts this to a Go time.Time.
func Date(dateStr string) string {
	return fmt.Sprintf("Date('%s')", dateStr)
}

// DateTime creates a DateTime marker string. Same as Date on the server side.
func DateTime(dateStr string) string {
	return fmt.Sprintf("DateTime('%s')", dateStr)
}

// Timestamp creates a Timestamp marker string. The server converts this to a Unix timestamp.
func Timestamp(unix any) string {
	return fmt.Sprintf("Timestamp('%v')", unix)
}

// ==================== Numeric Types ====================

// Boolean creates a Boolean marker. The server converts this to a Go bool.
func Boolean(val bool) string {
	return fmt.Sprintf("Boolean(%t)", val)
}

// Integer creates an Integer marker. The server converts this to int32.
func Integer(n any) string {
	return fmt.Sprintf("Integer(%v)", n)
}

// Bigint creates a Bigint marker. The server converts this to int64.
func Bigint(n any) string {
	return fmt.Sprintf("Bigint(%v)", n)
}

// Float creates a Float marker. The server converts this to float32.
func Float(n any) string {
	return fmt.Sprintf("Float(%v)", n)
}

// Double creates a Double marker. The server converts this to float64.
func Double(n any) string {
	return fmt.Sprintf("Double(%v)", n)
}

// ==================== Structured Types ====================

// Array creates an Array marker. The server converts this to a Go slice.
func Array(arr []any) string {
	jsonBytes, _ := json.Marshal(arr)
	return fmt.Sprintf("Array(%s)", string(jsonBytes))
}

// Map creates a Map marker. The server converts this to map[string]any.
func Map_(obj map[string]any) string {
	jsonBytes, _ := json.Marshal(obj)
	return fmt.Sprintf("Map(%s)", string(jsonBytes))
}

// MapArray creates a Map[] marker. The server converts this to []map[string]any.
func MapArray(arr []map[string]any) string {
	jsonBytes, _ := json.Marshal(arr)
	return fmt.Sprintf("Map[](%s)", string(jsonBytes))
}

// ==================== Null Types ====================

// Null markers that the server converts to Go nil.
const (
	Null      = "Null"
	None      = "None"
	Nil       = "Nil"
	Empty     = ""
	UNDEFINED = "UNDEFINED"
	Undefined = "Undefined"
)

// ==================== Time Helpers ====================

// Time helper markers that the server evaluates at request time.
// These are string values that, when sent to the server, are replaced
// with the actual current time on the server side via Go's time.Now().
const (
	TimeNowMarker               = "time.Now()"
	TimeNowUnixMarker           = "time.Now().Unix()"
	TimeNowUnixNanoMarker       = "time.Now().UnixNano()"
	TimeNowWeekdayMarker        = "time.Now().Weekday()"
	TimeNowWeekdayChineseMarker = "time.Now().Weekday().Chinese()"
)

// timeNow is a convenience struct grouping all time-now helpers.
var TimeNow = struct {
	Now               func() string
	NowUnix           func() string
	NowUnixNano       func() string
	NowWeekday        func() string
	NowWeekdayChinese func() string
}{
	Now:               func() string { return TimeNowMarker },
	NowUnix:           func() string { return TimeNowUnixMarker },
	NowUnixNano:       func() string { return TimeNowUnixNanoMarker },
	NowWeekday:        func() string { return TimeNowWeekdayMarker },
	NowWeekdayChinese: func() string { return TimeNowWeekdayChineseMarker },
}
