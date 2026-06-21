package t1y

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	json "github.com/goccy/go-json"
)

// T1YOS is the main client for the t1yOS Serverless Platform SDK.
//
// It provides:
//   - Initialization with server time sync
//   - Chainable database operations via DB.Collection(name)
//   - Cloud function invocation
//   - Metadata retrieval
//   - Cryptographic utilities
type T1YOS struct {
	config *internalConfig
	client *http.Client

	// DB provides chainable database operations.
	DB *DB
}

// DB provides access to database operations.
type DB struct {
	client *T1YOS
}

// Collection returns a T1Collection for the given collection name.
func (db *DB) Collection(name string) *T1Collection {
	return newCollection(db.client, name)
}

// ToObjectID converts a 24-char hex string to an ObjectID marker.
func (db *DB) ToObjectID(id string) (string, error) {
	if err := assertObjectID(id); err != nil {
		return "", err
	}
	return fmt.Sprintf("ObjectID('%s')", id), nil
}

// GetCollections returns all collections in the application's database.
func (db *DB) GetCollections(ctx context.Context) (*ApiResponse[CollectionsResult], error) {
	return request[CollectionsResult](db.client, ctx, http.MethodGet, "/v5/schemas", nil, db.client.config.isSafeMode)
}

// NewClient creates a new T1YOS client with the given configuration.
//
// Required parameters:
//   - AppID: application ID (must be >= 1001)
//   - APIKey: 32-character API key
//   - SecretKey: 32-character secret key
//
// Optional parameters (with defaults):
//   - BaseURL: default "https://myapp.t1y.net"
//   - Version: default 0
//   - IsSafeMode: default false
//   - TimeFormat: default "YYYY-MM-DD HH:mm:ss"
//   - Offset: default 0
func NewClient(config *Config) (*T1YOS, error) {
	if config == nil {
		return nil, NewValidationError("config cannot be nil")
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	internalCfg := config.resolveDefaults()

	c := &T1YOS{
		config: internalCfg,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}

	c.DB = &DB{client: c}

	return c, nil
}

// Init initializes the SDK by syncing with the server.
//
// Calls GET /init/:appId to:
// 1. Get the server's current UTC Unix timestamp
// 2. Get the server's isSafeMode setting
//
// The time offset is computed as: server.unix - client.unix
// This offset is used for all subsequent request signing to prevent clock skew issues.
func (c *T1YOS) Init(ctx context.Context) error {
	path := fmt.Sprintf("/init/%d", c.config.appID)
	resp, err := request[InitResult](c, ctx, http.MethodGet, path, nil, false)
	if err != nil {
		// Graceful degradation: use defaults
		c.config.isSafeMode = false
		c.config.offset = 0
		return fmt.Errorf("failed to get time offset from server, defaulting to 0: %w", err)
	}

	c.config.isSafeMode = resp.Data.IsSafeMode
	c.config.offset = resp.Data.Unix - time.Now().UTC().Unix()
	return nil
}

// ==================== Public API ====================

// GetMeta returns application metadata.
//
// If field is empty, all metadata is returned as key-value pairs.
// If field is specified, a specific metadata field value is returned.
func (c *T1YOS) GetMeta(ctx context.Context, field string) (*ApiResponse[MetaResults], error) {
	queryPath := ""
	if field != "" {
		queryPath = fmt.Sprintf("?field=%s", url.QueryEscape(field))
	}
	return request[MetaResults](c, ctx, http.MethodGet, "/v5/meta"+queryPath, nil, c.config.isSafeMode)
}

// GetMetaField returns a specific metadata field value.
func (c *T1YOS) GetMetaField(ctx context.Context, field string) (*ApiResponse[MetaResult], error) {
	if field == "" {
		return nil, NewValidationError("Meta field must be a non-empty string")
	}
	queryPath := fmt.Sprintf("?field=%s", url.QueryEscape(field))
	return request[MetaResult](c, ctx, http.MethodGet, "/v5/meta"+queryPath, nil, c.config.isSafeMode)
}

// CheckUpdate checks if there's a newer version of the application available.
//
// Compares the server's version metadata field against the configured version.
// Returns true if the server version is greater than the client version.
func (c *T1YOS) CheckUpdate(ctx context.Context) (bool, error) {
	resp, err := c.GetMetaField(ctx, "version")
	if err != nil {
		return false, err
	}

	version, ok := resp.Data.Result.(float64)
	if !ok {
		return false, nil
	}

	return int(version) > c.config.version, nil
}

// CallFunc calls a cloud function (.jsc file).
//
// If name doesn't end with .jsc, it's auto-appended.
// If name ends with /, index.jsc is appended.
// If name ends with .js, it's replaced with .jsc.
//
// The enableSafeMode parameter overrides safe mode for this call.
// Pass nil to use the client's default setting.
func (c *T1YOS) CallFunc(ctx context.Context, name string, params any, enableSafeMode *bool) (*ApiResponse[any], error) {
	if name == "" {
		return nil, NewValidationError("Function name must be a non-empty string")
	}

	safeMode := c.config.isSafeMode
	if enableSafeMode != nil {
		safeMode = *enableSafeMode
	}

	path := fmt.Sprintf("/%d/%s", c.config.appID, EnsureJscExtension(name))
	return request[any](c, ctx, http.MethodPost, path, params, safeMode)
}

// Request executes an authenticated HTTP request with full signing and encryption.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE)
//   - path: API path (e.g., /v5/classes/users)
//   - params: Request parameters/body
//   - encryption: Override encryption (if nil, defaults to client.isSafeMode)
func (c *T1YOS) Request(ctx context.Context, method, path string, params any, encryption *bool) (*ApiResponse[any], error) {
	safeMode := c.config.isSafeMode
	if encryption != nil {
		safeMode = *encryption
	}

	return request[any](c, ctx, method, path, params, safeMode)
}

// ==================== Utility Methods ====================

// AssertObjectID validates an ObjectID hex string.
// Returns an error if invalid, nil if valid.
func (c *T1YOS) AssertObjectID(idStr string) error {
	return assertObjectID(idStr)
}

// IsNonEmptyObject checks if a value is a non-nil, non-array map with at least one key.
func (c *T1YOS) IsNonEmptyObject(value any) bool {
	return IsNonEmptyObject(value)
}

// IsPlainObject checks if a value is a non-nil, non-array map.
func (c *T1YOS) IsPlainObject(value any) bool {
	return IsPlainObject(value)
}

// IsNonEmptyArrayWithNonEmptyObjects checks if a value is a non-empty slice
// where every element is a non-empty map.
func (c *T1YOS) IsNonEmptyArrayWithNonEmptyObjects(value any) bool {
	return IsNonEmptyArrayWithNonEmptyObjects(value)
}

// HMACSHA256 computes HMAC-SHA256 and returns the hex digest.
func (c *T1YOS) HMACSHA256(secret, message string) string {
	return HMACSHA256(secret, message)
}

// VerifyHMACSHA256 verifies an HMAC-SHA256 signature using constant-time comparison.
func (c *T1YOS) VerifyHMACSHA256(secret, message, signature string) bool {
	return VerifyHMACSHA256(secret, message, signature)
}

// ==================== Internal: HTTP Request ====================

// request executes an HTTP request with full authentication and encryption.
func request[T any](c *T1YOS, ctx context.Context, method, path string, params any, useSafeMode bool) (*ApiResponse[T], error) {
	// Normalize base URL
	baseURL := NormalizeBaseURL(c.config.baseURL)

	// Convert date types in params
	convertedParams := convertDateTypes(params)

	// Build request body
	var bodyBytes []byte
	var rawBodyString string

	if method != http.MethodGet && convertedParams != nil {
		jsonBody, err := json.Marshal(convertedParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		if useSafeMode {
			// Safe mode: encrypt the JSON body
			encryptedBody, err := EncryptAESGCM(jsonBody, []byte(c.config.secretKey))
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt request body: %w", err)
			}
			bodyBytes = []byte(encryptedBody)
			rawBodyString = encryptedBody
		} else {
			bodyBytes = jsonBody
			rawBodyString = string(jsonBody)
		}
	}

	// Build URL
	fullURL, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// For GET requests with params, append as query string
	if method == http.MethodGet && convertedParams != nil {
		if paramsMap, ok := convertedParams.(map[string]any); ok && len(paramsMap) > 0 {
			appendQueryParams(fullURL, paramsMap)
		}
	}

	// Compute timestamp with offset
	timestamp := time.Now().UTC().Unix() + c.config.offset

	// Get path + query for signing
	pathAndQuery := fullURL.Path
	if fullURL.RawQuery != "" {
		pathAndQuery += "?" + fullURL.RawQuery
	}

	// Create HMAC-SHA256 signature
	sign := CreateSignature(SignatureInput{
		Method:       method,
		PathAndQuery: pathAndQuery,
		Body:         rawBodyString,
		AppID:        c.config.appID,
		Timestamp:    timestamp,
		SecretKey:    c.config.secretKey,
	})

	// Build HTTP request
	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth headers
	req.Header.Set("X-T1Y-Application-ID", fmt.Sprintf("%d", c.config.appID))
	req.Header.Set("X-T1Y-API-Key", c.config.apiKey)
	req.Header.Set("X-T1Y-Safe-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-T1Y-Safe-Sign", sign)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, handleHTTPError(err)
	}
	defer resp.Body.Close()

	// Process response
	return handleResponse[T](resp, useSafeMode, c.config.secretKey, c.config.timeFormat)
}

// handleResponse processes the HTTP response — decryption, timestamp formatting, and error wrapping.
func handleResponse[T any](resp *http.Response, isSafeMode bool, secretKey, timeFormat string) (*ApiResponse[T], error) {
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewT1YError(0, fmt.Sprintf("failed to read response body: %v", err), nil)
	}

	bodyStr := string(bodyBytes)

	// Determine if response is JSON
	contentType := resp.Header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json")

	// If safe mode is on and response is JSON, try to decrypt
	if isSafeMode && isJSON {
		decrypted, decryptErr := DecryptAESGCM(bodyStr, []byte(secretKey))
		if decryptErr == nil {
			bodyBytes = decrypted
			bodyStr = string(decrypted)
		}
		// If decryption fails, proceed with raw body
	}

	// Try to parse as standard ApiResponse
	if isJSON {
		var apiResp ApiResponse[T]
		if err := json.Unmarshal(bodyBytes, &apiResp); err == nil {
			// Format timestamps in data
			var formatted any = formatTimestampsToLocal(apiResp.Data, timeFormat)
			if typed, ok := formatted.(T); ok {
				apiResp.Data = typed
			}

			// If not a 2xx status, return as error
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, NewT1YError(apiResp.Code, apiResp.Message, apiResp.Data)
			}

			return &apiResp, nil
		}
	}

	// Non-standard response handling
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success but not standard format — wrap it
		var zero T
		return &ApiResponse[T]{
			Code:    0,
			Message: "ok",
			Data:    zero,
		}, nil
	}

	return nil, NewT1YError(resp.StatusCode, resp.Status, bodyStr)
}

// handleHTTPError wraps HTTP client errors as T1YError.
func handleHTTPError(err error) error {
	if err == nil {
		return nil
	}

	// Check for context deadline exceeded (timeout)
	if err == context.DeadlineExceeded {
		return NewT1YError(408, "Request timeout", nil)
	}

	// Check for context canceled
	if err == context.Canceled {
		return NewT1YError(499, "Request canceled", nil)
	}

	return NewT1YError(0, err.Error(), nil)
}
