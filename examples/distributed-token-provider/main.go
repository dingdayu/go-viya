package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dingdayu/go-viya"
)

type sharedTokenCache interface {
	AccessToken(ctx context.Context) (string, error)
}

type distributedTokenProvider struct {
	cache sharedTokenCache
}

func (p distributedTokenProvider) Token(ctx context.Context) (string, error) {
	token, err := p.cache.AccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("viya access token: %w", err)
	}
	if token == "" {
		return "", viya.ErrViyaAuthFailed
	}
	return token, nil
}

type cachedAccessToken struct {
	token     string
	expiresAt time.Time
}

type tenantTokenCache struct {
	current cachedAccessToken
}

func (c tenantTokenCache) AccessToken(context.Context) (string, error) {
	if c.current.token != "" && time.Until(c.current.expiresAt) > time.Minute {
		return c.current.token, nil
	}
	return c.refreshWithDistributedLock()
}

func (c tenantTokenCache) refreshWithDistributedLock() (string, error) {
	// In production, refresh through a single owner protected by a distributed
	// lock, and keep refresh tokens in a secret manager.
	return "", viya.ErrViyaAuthFailed
}

func main() {
	ctx := context.Background()
	baseURL := "https://viya.example.com"

	cache := tenantTokenCache{
		current: cachedAccessToken{
			token:     "access-token-from-shared-cache",
			expiresAt: time.Now().Add(15 * time.Minute),
		},
	}

	provider := distributedTokenProvider{cache: cache}
	client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(provider))

	users, err := client.GetIdentitiesUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("users: %d", users.Count)
}
