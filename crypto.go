package t1y

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	json "github.com/goccy/go-json"
)

// ==================== AES-256-GCM ====================

// AESGCMPayload represents the encrypted payload structure.
// Matches the Go server's AESGCMEncryptedPayload struct.
type AESGCMPayload struct {
	Nonce string `json:"n"`
	Data  string `json:"j"`
	Tag   string `json:"t"`
}

// EncryptAESGCM encrypts data using AES-256-GCM.
// The key must be exactly 32 bytes.
// Returns a JSON string of { n, j, t } payload.
func EncryptAESGCM(data []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key length must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nil, nonce, data, nil)
	tagLen := 16
	ciphertext := sealed[:len(sealed)-tagLen]
	tag := sealed[len(sealed)-tagLen:]

	payload := AESGCMPayload{
		Nonce: base64.StdEncoding.EncodeToString(nonce),
		Data:  base64.StdEncoding.EncodeToString(ciphertext),
		Tag:   base64.StdEncoding.EncodeToString(tag),
	}

	result, _ := json.Marshal(payload)
	return string(result), nil
}

// DecryptAESGCM decrypts data using AES-256-GCM.
// jsonPayload is a JSON string of { n, j, t } payload.
// The key must be exactly 32 bytes.
// Returns the decrypted plaintext.
func DecryptAESGCM(jsonPayload string, key []byte) (plaintext []byte, err error) {
	if len(key) != 32 {
		return nil, errors.New("key length must be 32 bytes")
	}

	var payload AESGCMPayload
	if err = json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
		return nil, fmt.Errorf("invalid json payload: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce base64: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext base64: %w", err)
	}
	tag, err := base64.StdEncoding.DecodeString(payload.Tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag base64: %w", err)
	}

	if len(nonce) != 12 {
		return nil, fmt.Errorf("invalid nonce length: %d (expected 12)", len(nonce))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cipher init failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm init failed: %w", err)
	}

	sealed := append(ciphertext, tag...)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("gcm open panic: %v", r)
		}
	}()

	plaintext, err = gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("gcm open failed: %w", err)
	}

	return plaintext, nil
}

// ==================== SHA-256 ====================

// SHA256Hex computes the SHA-256 hash of a string and returns the hex digest.
func SHA256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// ==================== HMAC-SHA256 ====================

// HMACSHA256 computes HMAC-SHA256 and returns the hex digest.
func HMACSHA256(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA256 verifies an HMAC-SHA256 signature using constant-time comparison.
func VerifyHMACSHA256(secret, message, signature string) bool {
	expected := HMACSHA256(secret, message)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ==================== Request Signing ====================

// SignatureInput contains the parameters for creating a request signature.
type SignatureInput struct {
	// HTTP method in uppercase (GET, POST, PUT, DELETE).
	Method string

	// URL path including query string, e.g. /v5/classes/users?field=name.
	PathAndQuery string

	// Request body as a raw string (empty string for GET requests).
	Body string

	// Application ID.
	AppID int

	// 10-digit Unix timestamp.
	Timestamp int64

	// 32-character Secret Key.
	SecretKey string
}

// CreateSignature creates an HMAC-SHA256 signature for a T1Y API request.
//
// The message format is (each line separated by \n):
//  1. HTTP method (uppercase)
//  2. URL path + query string
//  3. SHA-256 hex digest of the request body
//  4. Application ID (as string)
//  5. Unix timestamp (as string)
//
// Returns a 64-character hex-encoded HMAC-SHA256 signature.
func CreateSignature(input SignatureInput) string {
	bodyHash := SHA256Hex(input.Body)

	message := fmt.Sprintf("%s\n%s\n%s\n%d\n%d",
		input.Method,
		input.PathAndQuery,
		bodyHash,
		input.AppID,
		input.Timestamp,
	)

	return HMACSHA256(input.SecretKey, message)
}

// GetSafeTimestamp returns the current UTC Unix timestamp adjusted by the given offset.
// Returns a 10-digit Unix timestamp string.
func GetSafeTimestamp(offset int64) string {
	return fmt.Sprintf("%d", time.Now().UTC().Unix()+offset)
}
