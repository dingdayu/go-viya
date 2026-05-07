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
	serverID := mustEnv("VIYA_CAS_SERVER")
	caslib := mustEnv("VIYA_CASLIB")
	table := mustEnv("VIYA_CASTABLE")
	scope := envDefault("VIYA_CAS_SCOPE", "global")

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := viya.NewClientCredentialsTokenProvider(clientID, clientSecret, baseURL)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, viya.WithBaseURL(baseURL), viya.WithTokenProvider(tokens))

	if err := client.LoadCASTableToMemory(ctx, serverID, caslib, table, true, scope); err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %s:%s.%s", serverID, caslib, table)

	if os.Getenv("VIYA_CAS_UNLOAD") == "1" {
		if err := client.UnloadCASTableFromMemory(ctx, serverID, caslib, table); err != nil {
			log.Fatal(err)
		}
		log.Printf("unloaded %s:%s.%s", serverID, caslib, table)
	}
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
