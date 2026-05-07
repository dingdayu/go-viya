package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/dingdayu/go-viya"
)

func main() {
	ctx := context.Background()

	baseURL := mustEnv("VIYA_BASE_URL")
	clientID := mustEnv("VIYA_CLIENT_ID")
	clientSecret := mustEnv("VIYA_CLIENT_SECRET")
	definitionName := envDefault("VIYA_CONFIGURATION_DEFINITION", "sas.identities.providers.ldap.user")

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	tokens, err := viya.NewClientCredentialsTokenProvider(clientID, clientSecret, u)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, viya.WithBaseURL(u), viya.WithTokenProvider(tokens))

	body, err := client.GetConfiguration(ctx, definitionName)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(body)
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}

func envDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
