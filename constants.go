package t1y

import "time"

// Default base URL for the t1yOS platform.
const DefaultBaseURL = "https://myapp.t1y.net"

// Minimum valid application ID.
const MinAppID = 1001

// Required length for API Key.
const APIKeyLength = 32

// Required length for Secret Key.
const SecretKeyLength = 32

// Default application version.
const DefaultVersion = 0

// Default time format for createdAt/updatedAt fields.
const DefaultTimeFormat = "YYYY-MM-DD HH:mm:ss"

// Default time offset in seconds.
const DefaultOffset = 0

// Default safe mode setting.
const DefaultSafeMode = false

// Maximum time difference allowed for request timestamp (seconds).
const MaxTimeDiff = 10

// Request timeout (5 minutes).
const RequestTimeout = 5 * time.Minute

// Maximum page size for find queries.
const MaxPageSize = 100

// Default page size.
const DefaultPageSize = 10

// ObjectID hex string length.
const ObjectIDLength = 24

// API version prefix.
const APIVersion = "v5"
