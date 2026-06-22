package t1y

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	json "github.com/goccy/go-json"
)

// ==================== Validators ====================

// validateAppID validates that the application ID is a valid integer >= MinAppID.
func validateAppID(appID int) error {
	if appID < MinAppID {
		return NewValidationError(fmt.Sprintf("appId must be >= %d", MinAppID))
	}
	return nil
}

// validateAPIKey validates that the API Key is exactly the required length.
func validateAPIKey(apiKey string) error {
	if len(apiKey) != APIKeyLength {
		return NewValidationError(
			fmt.Sprintf("apiKey must be exactly %d characters (got %d)", APIKeyLength, len(apiKey)),
		)
	}
	return nil
}

// validateSecretKey validates that the Secret Key is exactly the required length.
func validateSecretKey(secretKey string) error {
	if len(secretKey) != SecretKeyLength {
		return NewValidationError(
			fmt.Sprintf("secretKey must be exactly %d characters (got %d)", SecretKeyLength, len(secretKey)),
		)
	}
	return nil
}

// validateBaseURL validates the base URL format.
func validateBaseURL(baseURL string) error {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return NewValidationError(`baseUrl must start with "http://" or "https://"`)
	}
	return nil
}

// validateConfig validates all configuration parameters at once.
func validateConfig(config *Config) error {
	if config.BaseURL != "" {
		if err := validateBaseURL(config.BaseURL); err != nil {
			return err
		}
	}
	if err := validateAppID(config.AppID); err != nil {
		return err
	}
	if err := validateAPIKey(config.APIKey); err != nil {
		return err
	}
	if err := validateSecretKey(config.SecretKey); err != nil {
		return err
	}
	if config.Version < 0 {
		return NewValidationError("version must be a non-negative integer")
	}
	return nil
}

// assertObjectID validates an ObjectID hex string.
// Returns nil if valid, error otherwise.
func assertObjectID(idStr string) error {
	if !objectIDPattern.MatchString(idStr) {
		return NewValidationError(fmt.Sprintf("Invalid ObjectID string: \"%s\"", idStr))
	}
	return nil
}

// ==================== Object Utilities ====================

// IsNonEmptyObject checks if a value is a non-nil, non-array map with at least one key.
func IsNonEmptyObject(value any) bool {
	if value == nil {
		return false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Map {
		return false
	}
	return rv.Len() > 0
}

// IsPlainObject checks if a value is a non-nil, non-array map.
func IsPlainObject(value any) bool {
	if value == nil {
		return false
	}
	rv := reflect.ValueOf(value)
	return rv.Kind() == reflect.Map
}

// IsNonEmptyArrayWithNonEmptyObjects checks if a value is a non-empty slice
// where every element is a non-empty map.
func IsNonEmptyArrayWithNonEmptyObjects(value any) bool {
	if value == nil {
		return false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice || rv.Len() == 0 {
		return false
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() != reflect.Map || elem.Len() == 0 {
			return false
		}
	}
	return true
}

// ==================== Date Conversion ====================

// convertDateTypes recursively converts Go time.Time values and large timestamp
// numbers into marker strings that the server's GetDataTypes() recognizes.
//
// - time.Time → Date('RFC3339Nano')
// - int64/int >= 10 digits → Timestamp('unix')
// - Already-marker strings are passed through.
func convertDateTypes(value any) any {
	if value == nil {
		return nil
	}

	// Handle time.Time
	if t, ok := value.(time.Time); ok {
		return fmt.Sprintf("Date('%s')", t.Format(time.RFC3339Nano))
	}

	rv := reflect.ValueOf(value)

	// Handle numbers: int with >= 10 digits that falls within a valid
	// Unix timestamp range (seconds or milliseconds) → Timestamp marker.
	// This avoids misidentifying non-timestamp large integers (e.g. user IDs).
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := rv.Int()
		if isUnixTimestamp(v) {
			return fmt.Sprintf("Timestamp('%d')", v)
		}
		return v
	case reflect.Float64:
		v := rv.Float()
		if isUnixTimestampFloat(v) {
			return fmt.Sprintf("Timestamp('%v')", v)
		}
		return v
	}

	// Handle slices
	if rv.Kind() == reflect.Slice {
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = convertDateTypes(rv.Index(i).Interface())
		}
		return result
	}

	// Handle maps
	if rv.Kind() == reflect.Map {
		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			result[fmt.Sprintf("%v", key.Interface())] = convertDateTypes(rv.MapIndex(key).Interface())
		}
		return result
	}

	return value
}

// ==================== URL Utilities ====================

// NormalizeBaseURL strips trailing slashes from a base URL.
func NormalizeBaseURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}

// appendQueryParams appends query parameters from a map to a URL.
// Skips nil values. Non-string values are JSON-marshaled.
func appendQueryParams(u *url.URL, params map[string]any) {
	q := u.Query()
	for key, value := range params {
		if value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			q.Set(key, v)
		default:
			jsonBytes, _ := json.Marshal(v)
			q.Set(key, string(jsonBytes))
		}
	}
	u.RawQuery = q.Encode()
}

// ==================== Time Formatting ====================

// formatTimestampsToLocal recursively converts all createdAt and updatedAt fields
// in a data structure from UTC strings to local time formatted strings.
func formatTimestampsToLocal(data any, format string) any {
	return traverseTimestamps(data, format)
}

func traverseTimestamps(value any, format string) any {
	if value == nil {
		return nil
	}

	// Handle slices
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Slice {
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = traverseTimestamps(rv.Index(i).Interface(), format)
		}
		return result
	}

	// Handle maps
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Map {
		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if keyStr == "createdAt" || keyStr == "updatedAt" {
				if strVal, ok := rv.MapIndex(key).Interface().(string); ok {
					result[keyStr] = formatLocalTime(strVal, format)
					continue
				}
			}
			result[keyStr] = traverseTimestamps(rv.MapIndex(key).Interface(), format)
		}
		return result
	}

	return value
}

// formatLocalTime formats a UTC date string to local time using the given format template.
//
// Format tokens:
//   - YYYY: 4-digit year
//   - MM: 2-digit month (01-12)
//   - DD: 2-digit day (01-31)
//   - HH: 2-digit hour (00-23)
//   - mm: 2-digit minute (00-59)
//   - ss: 2-digit second (00-59)
func formatLocalTime(utcString, format string) string {
	t, err := time.Parse(time.RFC3339Nano, utcString)
	if err != nil {
		// Try other common formats
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02T15:04:05Z",
		}
		parsed := false
		for _, layout := range layouts {
			if t, err = time.Parse(layout, utcString); err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return utcString // Return original if parsing fails
		}
	}

	local := t.Local()

	result := format
	result = strings.ReplaceAll(result, "YYYY", fmt.Sprintf("%04d", local.Year()))
	result = strings.ReplaceAll(result, "MM", fmt.Sprintf("%02d", local.Month()))
	result = strings.ReplaceAll(result, "DD", fmt.Sprintf("%02d", local.Day()))
	result = strings.ReplaceAll(result, "HH", fmt.Sprintf("%02d", local.Hour()))
	result = strings.ReplaceAll(result, "mm", fmt.Sprintf("%02d", local.Minute()))
	result = strings.ReplaceAll(result, "ss", fmt.Sprintf("%02d", local.Second()))

	return result
}

// ==================== JSC Extension Helper ====================

// EnsureJscExtension ensures a function name has the .jsc extension.
// Rules:
//   - If name doesn't end with .jsc, it's auto-appended
//   - If name ends with /, index.jsc is appended
//   - If name ends with .js, it's replaced with .jsc
func EnsureJscExtension(input string) string {
	// Normalize: remove leading slash
	path := input
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// Separate hash fragment
	hashIndex := strings.Index(path, "#")
	hash := ""
	if hashIndex != -1 {
		hash = path[hashIndex:]
		path = path[:hashIndex]
	}

	// Separate query string
	qIndex := strings.Index(path, "?")
	query := ""
	if qIndex != -1 {
		query = path[qIndex:]
		path = path[:qIndex]
	}

	// Apply extension rules
	if strings.HasSuffix(path, "/") {
		path = path + "index.jsc"
	} else if strings.HasSuffix(path, ".jsc") {
		// Already has .jsc
	} else if strings.HasSuffix(path, ".js") {
		path = strings.TrimSuffix(path, ".js") + ".jsc"
	} else {
		path = path + ".jsc"
	}

	return path + query + hash
}

// ==================== Timestamp Heuristic ====================

// isUnixTimestamp checks if an int64 value falls within a valid Unix timestamp
// range (seconds: ~2001–2099, or milliseconds: ~2001–2099). This prevents
// misidentifying large non-timestamp integers (e.g. user IDs, phone numbers)
// as timestamps.
func isUnixTimestamp(v int64) bool {
	return (v >= MinTimestampSeconds && v <= MaxTimestampSeconds) ||
		(v >= MinTimestampMilliseconds && v <= MaxTimestampMilliseconds)
}

// isUnixTimestampFloat checks if a float64 value falls within a valid Unix
// timestamp range (seconds or milliseconds). Float timestamps may include a
// fractional second component.
func isUnixTimestampFloat(v float64) bool {
	return (v >= float64(MinTimestampSeconds) && v <= float64(MaxTimestampSeconds)) ||
		(v >= float64(MinTimestampMilliseconds) && v <= float64(MaxTimestampMilliseconds))
}
