package t1y

// Config holds the configuration for initializing the T1YOS client.
type Config struct {
	// Base URL of the t1yOS platform. Default: "https://myapp.t1y.net".
	BaseURL string `json:"baseUrl"`

	// Application ID. Required, must be an integer >= 1001.
	AppID int `json:"appId"`

	// API Key. Required, must be exactly 32 characters.
	APIKey string `json:"apiKey"`

	// Secret Key. Required, must be exactly 32 characters.
	SecretKey string `json:"secretKey"`

	// Application version. Default: 0.
	Version int `json:"version"`

	// Whether to enable safe mode (AES-256-GCM encryption). Default: false.
	IsSafeMode bool `json:"isSafeMode"`

	// Time format for createdAt/updatedAt fields. Default: "YYYY-MM-DD HH:mm:ss".
	TimeFormat string `json:"timeFormat"`

	// Time offset in seconds between client and server. Default: 0.
	Offset int64 `json:"offset"`
}

// internalConfig is the internal configuration with all optional fields resolved to defaults.
type internalConfig struct {
	baseURL    string
	appID      int
	apiKey     string
	secretKey  string
	version    int
	isSafeMode bool
	timeFormat string
	offset     int64
}

// resolveDefaults applies default values to any unset optional fields.
func (c *Config) resolveDefaults() *internalConfig {
	cfg := &internalConfig{
		appID:      c.AppID,
		apiKey:     c.APIKey,
		secretKey:  c.SecretKey,
		version:    c.Version,
		isSafeMode: c.IsSafeMode,
		timeFormat: c.TimeFormat,
		offset:     c.Offset,
	}

	if c.BaseURL == "" {
		cfg.baseURL = DefaultBaseURL
	} else {
		cfg.baseURL = c.BaseURL
	}

	if c.Version == 0 && !c.IsSafeMode {
		cfg.version = DefaultVersion
	}

	if c.TimeFormat == "" {
		cfg.timeFormat = DefaultTimeFormat
	}

	return cfg
}
