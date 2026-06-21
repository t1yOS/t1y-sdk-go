package t1y_test

import (
	"encoding/json"
	"strings"
	"testing"

	t1y "github.com/t1yOS/t1y-sdk-go"
)

// ==================== NewClient Tests ====================

func TestNewClientValid(t *testing.T) {
	client, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientNilConfig(t *testing.T) {
	_, err := t1y.NewClient(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestNewClientInvalidAppID(t *testing.T) {
	_, err := t1y.NewClient(&t1y.Config{
		AppID:     500,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
	})
	if err == nil {
		t.Fatal("expected error for invalid appId")
	}
}

func TestNewClientInvalidAPIKey(t *testing.T) {
	_, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    "short",
		SecretKey: strings.Repeat("b", 32),
	})
	if err == nil {
		t.Fatal("expected error for invalid apiKey")
	}
}

func TestNewClientInvalidSecretKey(t *testing.T) {
	_, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: "short",
	})
	if err == nil {
		t.Fatal("expected error for invalid secretKey")
	}
}

func TestNewClientInvalidBaseURL(t *testing.T) {
	_, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
		BaseURL:   "no-protocol.com",
	})
	if err == nil {
		t.Fatal("expected error for invalid baseUrl")
	}
}

func TestNewClientDefaultValues(t *testing.T) {
	client, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Defaults should be applied
	if err := client.AssertObjectID("507f1f77bcf86cd799439011"); err != nil {
		t.Fatalf("unexpected error validating ObjectID: %v", err)
	}
	if client.IsNonEmptyObject(map[string]any{"a": 1}) != true {
		t.Error("expected IsNonEmptyObject to return true for non-empty map")
	}
	if client.IsNonEmptyObject(map[string]any{}) != false {
		t.Error("expected IsNonEmptyObject to return false for empty map")
	}
	if client.IsNonEmptyObject(nil) != false {
		t.Error("expected IsNonEmptyObject to return false for nil")
	}
	if client.IsPlainObject(nil) != false {
		t.Error("expected IsPlainObject to return false for nil")
	}
	if client.IsPlainObject(map[string]any{}) != true {
		t.Error("expected IsPlainObject to return true for empty map")
	}
	if client.IsNonEmptyArrayWithNonEmptyObjects(nil) != false {
		t.Error("expected IsNonEmptyArrayWithNonEmptyObjects to return false for nil")
	}
}

func TestDBCollectionAccessor(t *testing.T) {
	client, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	coll := client.DB.Collection("users")
	if coll == nil {
		t.Fatal("expected non-nil collection")
	}
}

func TestDBToObjectID(t *testing.T) {
	client, err := t1y.NewClient(&t1y.Config{
		AppID:     1001,
		APIKey:    strings.Repeat("a", 32),
		SecretKey: strings.Repeat("b", 32),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := client.DB.ToObjectID("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ObjectID('507f1f77bcf86cd799439011')" {
		t.Errorf("unexpected ObjectID: %s", result)
	}

	_, err = client.DB.ToObjectID("invalid")
	if err == nil {
		t.Fatal("expected error for invalid ObjectID")
	}
}

// ==================== Validator Tests ====================

func TestAssertObjectID(t *testing.T) {
	// Valid
	if err := (&t1y.T1YOS{}).AssertObjectID("507f1f77bcf86cd799439011"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid — too short
	err := (&t1y.T1YOS{}).AssertObjectID("short")
	if err == nil {
		t.Error("expected error for invalid ObjectID")
	}

	// Invalid — wrong chars
	err = (&t1y.T1YOS{}).AssertObjectID("gggggggggggggggggggggggg")
	if err == nil {
		t.Error("expected error for invalid ObjectID")
	}
}

// ==================== Crypto Tests ====================

func TestSHA256Hex(t *testing.T) {
	result := t1y.SHA256Hex("hello world")
	if len(result) != 64 {
		t.Errorf("expected 64-char hex, got %d", len(result))
	}

	// Test empty string (known vector)
	emptyHash := t1y.SHA256Hex("")
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if emptyHash != expected {
		t.Errorf("empty string hash mismatch:\n  got: %s\n  want: %s", emptyHash, expected)
	}

	// Test determinism
	a := t1y.SHA256Hex("test")
	b := t1y.SHA256Hex("test")
	if a != b {
		t.Error("sha256 should be deterministic")
	}
}

func TestHMACSHA256(t *testing.T) {
	result := t1y.HMACSHA256("secret", "message")
	if len(result) != 64 {
		t.Errorf("expected 64-char hex, got %d", len(result))
	}

	// Deterministic
	a := t1y.HMACSHA256("key", "data")
	b := t1y.HMACSHA256("key", "data")
	if a != b {
		t.Error("hmac should be deterministic")
	}

	// Different keys → different outputs
	c := t1y.HMACSHA256("key1", "data")
	d := t1y.HMACSHA256("key2", "data")
	if c == d {
		t.Error("different keys should produce different hmac")
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	secret := "secret"
	message := "message"
	sig := t1y.HMACSHA256(secret, message)

	if !t1y.VerifyHMACSHA256(secret, message, sig) {
		t.Error("should verify correct signature")
	}

	if t1y.VerifyHMACSHA256(secret, message, strings.Repeat("a", 64)) {
		t.Error("should reject incorrect signature")
	}
}

func TestEncryptDecryptAESGCMRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i)
	}

	plaintext := "Hello, T1Y!"
	encrypted, err := t1y.EncryptAESGCM([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify payload structure
	var payload map[string]string
	if err := json.Unmarshal([]byte(encrypted), &payload); err != nil {
		t.Fatalf("encrypted result is not valid JSON: %v", err)
	}
	if _, ok := payload["n"]; !ok {
		t.Error("payload missing 'n' field")
	}
	if _, ok := payload["j"]; !ok {
		t.Error("payload missing 'j' field")
	}
	if _, ok := payload["t"]; !ok {
		t.Error("payload missing 't' field")
	}

	decrypted, err := t1y.DecryptAESGCM(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", string(decrypted), plaintext)
	}
}

func TestEncryptAESGCMInvalidKey(t *testing.T) {
	_, err := t1y.EncryptAESGCM([]byte("test"), make([]byte, 16))
	if err == nil {
		t.Fatal("expected error for 16-byte key")
	}
}

func TestDecryptAESGCMInvalidKey(t *testing.T) {
	_, err := t1y.DecryptAESGCM(`{"n":"","j":"","t":""}`, make([]byte, 16))
	if err == nil {
		t.Fatal("expected error for 16-byte key")
	}
}

func TestEncryptAESGCMDifferentNonces(t *testing.T) {
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i)
	}

	enc1, _ := t1y.EncryptAESGCM([]byte("test"), key)
	enc2, _ := t1y.EncryptAESGCM([]byte("test"), key)

	// Different nonces should produce different ciphertexts
	if enc1 == enc2 {
		t.Error("encryptions with different nonces should differ")
	}

	// Both should decrypt correctly
	dec1, _ := t1y.DecryptAESGCM(enc1, key)
	dec2, _ := t1y.DecryptAESGCM(enc2, key)
	if string(dec1) != "test" || string(dec2) != "test" {
		t.Error("both should decrypt correctly")
	}
}

func TestCreateSignature(t *testing.T) {
	sig := t1y.CreateSignature(t1y.SignatureInput{
		Method:       "POST",
		PathAndQuery: "/v5/classes/users",
		Body:         `{"name":"Alice"}`,
		AppID:        1001,
		Timestamp:    1705312200,
		SecretKey:    strings.Repeat("a", 32),
	})

	if len(sig) != 64 {
		t.Errorf("expected 64-char hex, got %d", len(sig))
	}

	// Deterministic
	sig2 := t1y.CreateSignature(t1y.SignatureInput{
		Method:       "POST",
		PathAndQuery: "/v5/classes/users",
		Body:         `{"name":"Alice"}`,
		AppID:        1001,
		Timestamp:    1705312200,
		SecretKey:    strings.Repeat("a", 32),
	})
	if sig != sig2 {
		t.Error("signature should be deterministic")
	}

	// Different method → different signature
	sig3 := t1y.CreateSignature(t1y.SignatureInput{
		Method:       "GET",
		PathAndQuery: "/v5/classes/users",
		Body:         "",
		AppID:        1001,
		Timestamp:    1705312200,
		SecretKey:    strings.Repeat("a", 32),
	})
	if sig == sig3 {
		t.Error("different methods should produce different signatures")
	}
}

func TestGetSafeTimestamp(t *testing.T) {
	ts := t1y.GetSafeTimestamp(0)
	// Should be a 10-digit Unix timestamp string
	if len(ts) != 10 {
		t.Errorf("expected 10-digit timestamp, got %d digits: %s", len(ts), ts)
	}
}

// ==================== Special Types Tests ====================

func TestObjectID(t *testing.T) {
	result := t1y.ObjectID("507f1f77bcf86cd799439011")
	if result != "ObjectID('507f1f77bcf86cd799439011')" {
		t.Errorf("unexpected ObjectID: %s", result)
	}

	// Should panic on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid ObjectID")
		}
	}()
	t1y.ObjectID("invalid")
}

func TestDateHelpers(t *testing.T) {
	if result := t1y.Date("2024-01-15T10:30:00Z"); result != "Date('2024-01-15T10:30:00Z')" {
		t.Errorf("unexpected Date: %s", result)
	}
	if result := t1y.DateTime("2024-01-15T10:30:00Z"); result != "DateTime('2024-01-15T10:30:00Z')" {
		t.Errorf("unexpected DateTime: %s", result)
	}
	if result := t1y.Timestamp(1705312200); result != "Timestamp('1705312200')" {
		t.Errorf("unexpected Timestamp: %s", result)
	}
	if result := t1y.Timestamp("1705312200"); result != "Timestamp('1705312200')" {
		t.Errorf("unexpected Timestamp from string: %s", result)
	}
}

func TestNumericHelpers(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Boolean true", t1y.Boolean(true), "Boolean(true)"},
		{"Boolean false", t1y.Boolean(false), "Boolean(false)"},
		{"Integer", t1y.Integer(42), "Integer(42)"},
		{"Integer negative", t1y.Integer(-10), "Integer(-10)"},
		{"Bigint", t1y.Bigint(9007199254740991), "Bigint(9007199254740991)"},
		{"Float", t1y.Float(3.14), "Float(3.14)"},
		{"Double", t1y.Double(3.141592653589793), "Double(3.141592653589793)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %s, want %s", tt.got, tt.expected)
			}
		})
	}
}

func TestStructuredHelpers(t *testing.T) {
	if result := t1y.Array([]any{1, 2, 3}); result != "Array([1,2,3])" {
		t.Errorf("unexpected Array: %s", result)
	}
	if result := t1y.Map_(map[string]any{"key": "value"}); result != `Map({"key":"value"})` {
		t.Errorf("unexpected Map: %s", result)
	}
	if result := t1y.MapArray([]map[string]any{{"a": float64(1)}, {"b": float64(2)}}); result != `Map[]([{"a":1},{"b":2}])` {
		t.Errorf("unexpected MapArray: %s", result)
	}
}

func TestNullHelpers(t *testing.T) {
	if t1y.Null != "Null" {
		t.Error("unexpected Null value")
	}
	if t1y.None != "None" {
		t.Error("unexpected None value")
	}
	if t1y.Nil != "Nil" {
		t.Error("unexpected Nil value")
	}
	if t1y.Empty != "" {
		t.Error("unexpected Empty value")
	}
	if t1y.UNDEFINED != "UNDEFINED" {
		t.Error("unexpected UNDEFINED value")
	}
	if t1y.Undefined != "Undefined" {
		t.Error("unexpected Undefined value")
	}
}

func TestTimeNowHelpers(t *testing.T) {
	if result := t1y.TimeNow.Now(); result != "time.Now()" {
		t.Errorf("unexpected TimeNow.Now(): %s", result)
	}
	if result := t1y.TimeNow.NowUnix(); result != "time.Now().Unix()" {
		t.Errorf("unexpected TimeNow.NowUnix(): %s", result)
	}
	if result := t1y.TimeNow.NowUnixNano(); result != "time.Now().UnixNano()" {
		t.Errorf("unexpected TimeNow.NowUnixNano(): %s", result)
	}
	if result := t1y.TimeNow.NowWeekday(); result != "time.Now().Weekday()" {
		t.Errorf("unexpected TimeNow.NowWeekday(): %s", result)
	}
	if result := t1y.TimeNow.NowWeekdayChinese(); result != "time.Now().Weekday().Chinese()" {
		t.Errorf("unexpected TimeNow.NowWeekdayChinese(): %s", result)
	}
}

// ==================== Utility Tests ====================

func TestIsNonEmptyObject(t *testing.T) {
	if !t1y.IsNonEmptyObject(map[string]any{"a": 1}) {
		t.Error("should be true for non-empty map")
	}
	if t1y.IsNonEmptyObject(map[string]any{}) {
		t.Error("should be false for empty map")
	}
	if t1y.IsNonEmptyObject(nil) {
		t.Error("should be false for nil")
	}
	if t1y.IsNonEmptyObject([]any{1, 2, 3}) {
		t.Error("should be false for slice")
	}
}

func TestIsNonEmptyArrayWithNonEmptyObjects(t *testing.T) {
	if !t1y.IsNonEmptyArrayWithNonEmptyObjects([]map[string]any{{"a": 1}, {"b": 2}}) {
		t.Error("should be true for valid array")
	}
	if t1y.IsNonEmptyArrayWithNonEmptyObjects([]map[string]any{}) {
		t.Error("should be false for empty array")
	}
	if t1y.IsNonEmptyArrayWithNonEmptyObjects([]map[string]any{{}}) {
		t.Error("should be false for array with empty objects")
	}
	if t1y.IsNonEmptyArrayWithNonEmptyObjects(map[string]any{"a": 1}) {
		t.Error("should be false for non-slice")
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com///", "https://example.com"},
		{"https://example.com", "https://example.com"},
		{"http://localhost:8082/", "http://localhost:8082"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if result := t1y.NormalizeBaseURL(tt.input); result != tt.expected {
				t.Errorf("got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestEnsureJscExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello.jsc"},
		{"hello.jsc", "hello.jsc"},
		{"hello.js", "hello.jsc"},
		{"path/to/func", "path/to/func.jsc"},
		{"path/to/", "path/to/index.jsc"},
		{"hello?query=1", "hello.jsc?query=1"},
		{"/hello", "hello.jsc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if result := t1y.EnsureJscExtension(tt.input); result != tt.expected {
				t.Errorf("ensureJscExtension(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== Error Tests ====================

func TestT1YError(t *testing.T) {
	err := t1y.NewT1YError(403, "Forbidden", nil)
	if err.Code != 403 {
		t.Errorf("expected code 403, got %d", err.Code)
	}
	if err.Error() == "" {
		t.Error("expected non-empty error string")
	}
}

func TestValidationError(t *testing.T) {
	err := t1y.NewValidationError("something went wrong")
	if err.Error() == "" {
		t.Error("expected non-empty error string")
	}
}
