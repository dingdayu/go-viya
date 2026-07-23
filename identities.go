package viya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/codes"
)

// SAS Viya Identities API reference:
// https://developer.sas.com/rest-apis/identities

// RefreshIdentitiesCache triggers a refresh of the identities cache in SAS Viya.
func (c *Client) RefreshIdentitiesCache(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "RefreshIdentitiesCache")
	defer span.End()

	request := c.client.R().SetContext(ctx)
	resp, err := request.Post("/identities/cache/refreshes")
	if err != nil {
		return err
	}
	if !resp.IsStatusSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return fmt.Errorf("failed to refresh identities cache, status code: %d", resp.StatusCode())
	}
	return nil
}

// IdentitiesUsers describes a user entry returned by the SAS Viya Identities API.
type IdentitiesUsers struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProviderId  string `json:"providerId"`
	Type        string `json:"type"`
	Description string `json:"description"`
	State       string `json:"state"`
}

// IdentitiesUsersResp is a collection of SAS Viya identity users.
type IdentitiesUsersResp = ListResponse[IdentitiesUsers]

// GetIdentitiesUsers returns up to 100 SAS Viya identity users.
func (c *Client) GetIdentitiesUsers(ctx context.Context) (identitiesUserResp IdentitiesUsersResp, err error) {
	ctx, span := tracer.Start(ctx, "GetIdentitiesUsers")
	defer span.End()

	resp, err := c.client.R().SetHeader("Accept", "application/json").SetContext(ctx).SetResult(&identitiesUserResp).SetQueryParam("limit", "100").Get("/identities/users")
	if err != nil {
		return identitiesUserResp, err
	}
	if !resp.IsStatusSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return identitiesUserResp, fmt.Errorf("failed to get identities users, status code: %d", resp.StatusCode())
	}

	return identitiesUserResp, nil
}

// NOTE: The following identities helpers are not verified against a live SAS Viya deployment.
// LDAP group-based access control is recommended over manual objectFilter management
// since the provider has already switched to group-based configuration.

// GetIdentitiesLDAPUser retrieves the LDAP user provider configuration.
//
// The configuration service returns dynamic payloads, so the configuration is
// represented as map[string]any.
func (c *Client) GetIdentitiesLDAPUser(ctx context.Context) (map[string]any, error) {
	resp, err := c.GetConfiguration(ctx, "sas.identities.providers.ldap.user")
	if err != nil {
		return nil, err
	}

	var config ConfigurationsResp
	if err := json.Unmarshal([]byte(resp), &config); err != nil {
		return nil, fmt.Errorf("unmarshal configuration response: %w", err)
	}
	if len(config.Items) == 0 || config.Items[0] == nil {
		return nil, errors.New("configuration response has no items")
	}

	return config.Items[0], nil
}

// PatchIdentitiesLDAPUser updates the LDAP provider configuration with the supplied values.
func (c *Client) PatchIdentitiesLDAPUser(ctx context.Context, updates map[string]any) error {
	conf, err := c.GetIdentitiesLDAPUser(ctx)
	if err != nil {
		return err
	}

	// apply updates
	for k, v := range updates {
		conf[k] = v
	}

	// find the update link safely
	linksAny, ok := conf["links"]
	if !ok {
		return errors.New("configuration has no links")
	}
	linksSlice, ok := linksAny.([]any)
	if !ok {
		return errors.New("configuration links has unexpected type")
	}

	var link Link
	for _, linkItem := range linksSlice {
		linkMap, ok := linkItem.(map[string]any)
		if !ok {
			continue
		}
		if rel, _ := linkMap["rel"].(string); rel == "update" {
			link = Link{
				Href:   strOrEmpty(linkMap["href"]),
				Method: strOrEmpty(linkMap["method"]),
				Rel:    rel,
				URI:    strOrEmpty(linkMap["uri"]),
				Type:   strOrEmpty(linkMap["type"]),
			}
			break
		}
	}
	if link.Href == "" || link.Method == "" {
		return errors.New("update link not found in configuration")
	}

	req := c.client.R().SetContext(ctx)
	if link.Type != "" {
		req = req.SetContentType(link.Type)
	}

	response, err := req.SetBody(conf).Execute(link.Method, link.Href)
	if err != nil {
		return err
	}
	if !response.IsStatusSuccess() {
		return fmt.Errorf("failed to patch LDAP user, status code: %d", response.StatusCode())
	}
	return nil
}

// UpdateIdentitiesLDAPObjectFilter updates the LDAP object filter to include only the specified usernames.
func (c *Client) UpdateIdentitiesLDAPObjectFilter(ctx context.Context, usernames []string) error {
	var accountNames []string
	for _, username := range usernames {
		if username == "" {
			continue
		}
		accountNames = append(accountNames, fmt.Sprintf("(sAMAccountName=%s)", username))
	}

	if len(accountNames) == 0 {
		return errors.New("no usernames provided for LDAP object filter update")
	}

	updates := map[string]any{
		"objectFilter": fmt.Sprintf("(&(|%s)(objectClass=user))", strings.Join(accountNames, "")),
	}
	if err := c.PatchIdentitiesLDAPUser(ctx, updates); err != nil {
		return fmt.Errorf("updating LDAP object filter for %s: %w", strings.Join(usernames, ", "), err)
	}
	return nil
}

// strOrEmpty returns v as a string, or "" if v is not a string.
func strOrEmpty(v any) string {
	s, _ := v.(string)
	return s
}
