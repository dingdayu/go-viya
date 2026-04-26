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

	tokens, err := viya.NewClientCredentialsTokenProvider(baseURL, clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}

	viya.SetDefaultClient(viya.NewClient(ctx, baseURL, viya.WithTokenProvider(tokens)))

	client, err := viya.GetDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

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
