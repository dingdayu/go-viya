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
	username := mustEnv("VIYA_USERNAME")
	password := mustEnv("VIYA_PASSWORD")
	clientID := os.Getenv("VIYA_CLIENT_ID")
	clientSecret := os.Getenv("VIYA_CLIENT_SECRET")

	provider, err := viya.NewPasswordTokenProvider(
		baseURL,
		username,
		password,
		viya.WithOAuthClient(clientID, clientSecret),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(provider))

	users, err := client.GetIdentitiesUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("users: %d", users.Count)
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
