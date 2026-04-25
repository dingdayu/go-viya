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
func (c *Client) RefreshIdentitiesCache(ctx context.Context) (success bool, err error) {
	ctx, span := tracer.Start(ctx, "RefreshIdentitiesCache")
	defer span.End()

	request := c.client.R().SetContext(ctx)
	resp, err := request.Post("/identities/cache/refreshes")
	if err != nil {
		return false, err
	}
	if !resp.IsSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return false, fmt.Errorf("failed to refresh identities cache, status code: %d", resp.StatusCode())
	}
	return true, nil
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
	if !resp.IsSuccess() {
		span.SetStatus(codes.Error, resp.String())
		return identitiesUserResp, fmt.Errorf("failed to get identities users, status code: %d", resp.StatusCode())
	}

	return identitiesUserResp, nil
}

// NOTE: 以下未经过验证，推荐通过 AD 域中的组进行控制，objectFilter 已经切换到 通过组配置了

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

// PatchIdentitiesLDAPGroup updates the LDAP provider configuration with the supplied values.
//
// NOTE: Despite the method name, this updates the LDAP user provider configuration.
// The name is preserved for source compatibility.
func (c *Client) PatchIdentitiesLDAPGroup(ctx context.Context, updates map[string]any) (bool, error) {
	// NOTE: Despite the name, this updates the LDAP user provider configuration.
	// Keeping the name to avoid breaking external callers.
	conf, err := c.GetIdentitiesLDAPUser(ctx)
	if err != nil {
		return false, err
	}

	// apply updates
	for k, v := range updates {
		conf[k] = v
	}

	// find the update link safely
	linksAny, ok := conf["links"]
	if !ok {
		return false, errors.New("configuration has no links")
	}
	linksSlice, ok := linksAny.([]any)
	if !ok {
		return false, errors.New("configuration links has unexpected type")
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
		return false, errors.New("update link not found in configuration")
	}

	req := c.client.R().SetContext(ctx)
	if link.Type != "" {
		req = req.SetContentType(link.Type)
	}

	response, err := req.SetBody(conf).Execute(link.Method, link.Href)
	if err != nil {
		return false, err
	}
	return response.IsSuccess(), nil
}

// UpdateIdentitiesLDAPObjectFilter updates the LDAP object filter to include only the specified usernames.
func (c *Client) UpdateIdentitiesLDAPObjectFilter(ctx context.Context, usernames []string) (bool, error) {
	var accountNames []string
	for _, username := range usernames {
		if username == "" {
			continue
		}
		accountNames = append(accountNames, fmt.Sprintf("(sAMAccountName=%s)", username))
	}

	if len(accountNames) == 0 {
		return false, errors.New("no usernames provided for LDAP object filter update")
	}

	updates := map[string]any{
		"objectFilter": fmt.Sprintf("(&(|%s)(objectClass=user))", strings.Join(accountNames, "")),
	}
	success, err := c.PatchIdentitiesLDAPGroup(ctx, updates)
	if err != nil {
		return false, fmt.Errorf("updating LDAP object filter for %s: %w", strings.Join(usernames, ", "), err)
	}
	if !success {
		return false, fmt.Errorf("failed to update LDAP object filter for %s", strings.Join(usernames, ", "))
	}
	return true, nil
}

// strOrEmpty tries to assert v as string and returns "" if it can't.
func strOrEmpty(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
