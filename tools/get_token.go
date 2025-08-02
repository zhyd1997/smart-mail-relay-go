package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
	// "google.golang.org/api/option"
)

func main() {
	clientID := os.Getenv("GMAIL_CLIENT_ID")
	clientSecret := os.Getenv("GMAIL_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Please set GMAIL_CLIENT_ID and GMAIL_CLIENT_SECRET environment variables")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{gmail.GmailReadonlyScope, gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: %v\n", authURL)
	fmt.Println("\nAfter authorization, you'll be redirected to a URL. Copy the 'code' parameter from that URL.")

	var authCode string
	fmt.Print("\nEnter the authorization code: ")
	fmt.Scan(&authCode)

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	fmt.Printf("\nRefresh Token: %s\n", tok.RefreshToken)
	fmt.Printf("Access Token: %s\n", tok.AccessToken)
	fmt.Printf("Token Type: %s\n", tok.TokenType)
	fmt.Printf("Expiry: %v\n", tok.Expiry)

	fmt.Println("\nAdd the refresh token to your environment variables:")
	fmt.Printf("export GMAIL_REFRESH_TOKEN=\"%s\"\n", tok.RefreshToken)
}
