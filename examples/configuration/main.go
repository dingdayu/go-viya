package main

import (
	"context"
	"log"
	"os"

	"github.com/dingdayu/go-viya"
)

func main() {
	ctx := context.Background()

	baseURL := mustEnv("VIYA_BASE_URL")
	clientID := mustEnv("VIYA_CLIENT_ID")
	clientSecret := mustEnv("VIYA_CLIENT_SECRET")
	definitionName := envDefault("VIYA_CONFIGURATION_DEFINITION", "sas.identities.providers.ldap.user")

	tokens, err := viya.NewClientCredentialsTokenProvider(baseURL, clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(tokens))

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
