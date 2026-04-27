package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dingdayu/go-viya"
)

func newConfiguredClient(cfg cliConfig) (*viya.Client, context.Context, context.CancelFunc, cliConfig, error) {
	if err := cfg.loadDefaults(); err != nil {
		return nil, nil, nil, cfg, err
	}

	output, err := normalizeOutput(cfg.Output)
	if err != nil {
		return nil, nil, nil, cfg, err
	}
	cfg.Output = output

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	provider, err := cfg.tokenProvider()
	if err != nil {
		cancel()
		return nil, nil, nil, cfg, err
	}

	return viya.NewClient(ctx, cfg.BaseURL, viya.WithTokenProvider(provider)), ctx, cancel, cfg, nil
}

func (c *cliConfig) loadDefaults() error {
	home, _ := os.UserHomeDir()
	if c.SASConfigDir == "" && home != "" {
		c.SASConfigDir = filepath.Join(home, ".sas")
	}

	configValues := readJSONFile(filepath.Join(c.SASConfigDir, "config.json"))
	credentialValues := readJSONFile(filepath.Join(c.SASConfigDir, "credentials.json"))
	profile := firstNonEmpty(c.Profile, env("VIYA_PROFILE", "SAS_PROFILE"), lookupAny(configValues, "profile", "defaultProfile", "currentProfile"))

	c.BaseURL = firstNonEmpty(
		c.BaseURL,
		env("VIYA_BASE_URL", "SAS_VIYA_URL", "SAS_SERVICES_ENDPOINT", "SAS_BASE_URL"),
		lookupProfile(configValues, profile, "baseURL", "baseUrl", "endpoint", "url", "sasServicesEndpoint"),
		lookupProfile(credentialValues, profile, "baseURL", "baseUrl", "endpoint", "url", "sasServicesEndpoint"),
	)
	c.ClientID = firstNonEmpty(
		c.ClientID,
		env("VIYA_CLIENT_ID", "SAS_CLIENT_ID"),
		lookupProfile(configValues, profile, "clientID", "clientId", "client_id"),
		lookupProfile(credentialValues, profile, "clientID", "clientId", "client_id"),
		"go-viya",
	)
	c.ClientSecret = firstNonEmpty(
		c.ClientSecret,
		env("VIYA_CLIENT_SECRET", "SAS_CLIENT_SECRET"),
		lookupProfile(configValues, profile, "clientSecret", "client_secret"),
		lookupProfile(credentialValues, profile, "clientSecret", "client_secret"),
	)
	c.Username = firstNonEmpty(
		c.Username,
		env("VIYA_USERNAME", "SAS_USERNAME", "SAS_USER"),
		lookupProfile(credentialValues, profile, "username", "user", "userid", "userId"),
	)
	c.Password = firstNonEmpty(
		c.Password,
		env("VIYA_PASSWORD", "SAS_PASSWORD"),
		lookupProfile(credentialValues, profile, "password", "pass"),
	)
	c.AccessToken = firstNonEmpty(
		c.AccessToken,
		env("VIYA_ACCESS_TOKEN", "SAS_ACCESS_TOKEN", "ACCESS_TOKEN"),
		lookupProfile(credentialValues, profile, "accessToken", "access_token", "token"),
	)
	c.ContextID = firstNonEmpty(
		c.ContextID,
		env("VIYA_COMPUTE_CONTEXT_ID", "SAS_COMPUTE_CONTEXT_ID"),
		lookupProfile(configValues, profile, "computeContextId", "contextId"),
	)
	c.ContextName = firstNonEmpty(
		c.ContextName,
		env("VIYA_COMPUTE_CONTEXT_NAME", "SAS_COMPUTE_CONTEXT_NAME"),
		lookupProfile(configValues, profile, "computeContextName", "contextName"),
	)

	if c.BaseURL == "" {
		return fmt.Errorf("SAS Viya base URL is required")
	}
	return nil
}

func (c cliConfig) tokenProvider() (viya.TokenProvider, error) {
	if c.AccessToken != "" {
		return staticTokenProvider(c.AccessToken), nil
	}
	if c.ClientSecret != "" && c.Username == "" && c.Password == "" {
		return viya.NewClientCredentialsTokenProvider(c.BaseURL, c.ClientID, c.ClientSecret)
	}
	if c.Username != "" && c.Password != "" {
		return viya.NewPasswordTokenProvider(c.BaseURL, c.Username, c.Password, viya.WithOAuthClient(c.ClientID, c.ClientSecret))
	}
	return nil, fmt.Errorf("credentials are required; provide access token, client credentials, or username/password")
}

func readJSONFile(path string) map[string]any {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var value map[string]any
	if json.Unmarshal(content, &value) != nil {
		return nil
	}
	return value
}

func lookupProfile(values map[string]any, profile string, keys ...string) string {
	if len(values) == 0 {
		return ""
	}
	if value := lookupAny(values, keys...); value != "" {
		return value
	}
	if profile != "" {
		if nested, ok := nestedMap(values[profile]); ok {
			if value := lookupAny(nested, keys...); value != "" {
				return value
			}
		}
	}
	for _, container := range []string{"profiles", "contexts", "credentials"} {
		nested, ok := nestedMap(values[container])
		if !ok {
			continue
		}
		if profile != "" {
			if profileValues, ok := nestedMap(nested[profile]); ok {
				if value := lookupAny(profileValues, keys...); value != "" {
					return value
				}
			}
		}
		if profile == "" {
			for _, candidate := range nested {
				if profileValues, ok := nestedMap(candidate); ok {
					if value := lookupAny(profileValues, keys...); value != "" {
						return value
					}
				}
			}
		}
	}
	return ""
}

func lookupAny(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			if text := stringValue(value); text != "" {
				return text
			}
		}
	}
	return ""
}

func nestedMap(value any) (map[string]any, bool) {
	nested, ok := value.(map[string]any)
	return nested, ok
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

func env(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func requireFlag(name string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("--%s is required", name)
	}
	return nil
}
