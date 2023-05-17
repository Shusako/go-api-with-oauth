package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	discordOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8000/oauth2/callback/discord",
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		Scopes:       []string{"identify", "email", "guilds"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
)

func main() {
	http.HandleFunc("/login/discord", loginPage)
	http.HandleFunc("/oauth2/callback/discord", discordOauthCallback)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
