package daytona

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	api "github.com/PhilippBuschhaus/daytona-sdk-go/apiclient"
	"github.com/joho/godotenv"
)

// Config holds the SDK configuration
type Config struct {
	BaseURL    string
	APIKey     string
	OrgID      string // Organization ID for X-Daytona-Organization-ID header
	Timeout    time.Duration
	RetryCount int
}

// Client is the main SDK client
type Client struct {
	config    *Config
	apiClient *api.APIClient
	ctx       context.Context
}

// init loads .env file automatically when the package is imported
func init() {
	// Try to load .env file if it exists
	// Ignore error if file doesn't exist - environment variables might be set elsewhere
	_ = godotenv.Load()
	
	// Also try to load from parent directory (useful for examples)
	_ = godotenv.Load("../.env")
}

// NewClient creates a new Daytona SDK client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}

	// Set defaults from environment or hardcoded values
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("DAYTONA_API_URL")
		if config.BaseURL == "" {
			config.BaseURL = "app.daytona.io/api"
		}
	}

	if config.APIKey == "" {
		config.APIKey = os.Getenv("DAYTONA_API_KEY")
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required: set DAYTONA_API_KEY environment variable or provide it in config")
		}
	}

	if config.OrgID == "" {
		config.OrgID = os.Getenv("DAYTONA_ORG_ID")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create API configuration
	apiConfig := api.NewConfiguration()
	// Override the default server configuration
	apiConfig.Servers = api.ServerConfigurations{
		{
			URL: "https://app.daytona.io/api",
			Description: "Daytona API",
		},
	}
	// Also set operation servers for toolbox operations
	apiConfig.OperationServers = map[string]api.ServerConfigurations{
		"ToolboxAPIService.ExecuteCommand": {
			{URL: "https://app.daytona.io/api"},
		},
		"ToolboxAPIService.CreateFolder": {
			{URL: "https://app.daytona.io/api"},
		},
		"ToolboxAPIService.UploadFile": {
			{URL: "https://app.daytona.io/api"},
		},
		"ToolboxAPIService.DownloadFile": {
			{URL: "https://app.daytona.io/api"},
		},
	}
	apiConfig.AddDefaultHeader("Authorization", "Bearer "+config.APIKey)
	if config.OrgID != "" {
		apiConfig.AddDefaultHeader("X-Daytona-Organization-ID", config.OrgID)
	}

	// Configure HTTP client with custom timeout
	// Only set custom HTTP client if timeout is configured, otherwise use default
	if config.Timeout > 0 {
		httpClient := &http.Client{
			Timeout: config.Timeout,
		}
		apiConfig.HTTPClient = httpClient
	}

	// Create API client
	apiClient := api.NewAPIClient(apiConfig)
	
	return &Client{
		config:    config,
		apiClient: apiClient,
		ctx:       context.Background(),
	}, nil
}

// GetAPIClient returns the underlying API client
func (c *Client) GetAPIClient() *api.APIClient {
	return c.apiClient
}

// WithContext returns a new client with the specified context
func (c *Client) WithContext(ctx context.Context) *Client {
	return &Client{
		config:    c.config,
		apiClient: c.apiClient,
		ctx:       ctx,
	}
}