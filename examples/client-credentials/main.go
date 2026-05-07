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

	baseURLStr := mustEnv("VIYA_BASE_URL")
	clientID := mustEnv("VIYA_CLIENT_ID")
	clientSecret := mustEnv("VIYA_CLIENT_SECRET")

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := viya.NewClientCredentialsTokenProvider(clientID, clientSecret, baseURL)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, viya.WithBaseURL(baseURL), viya.WithTokenProvider(tokens))

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
