package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Keys struct {
	CLIENT_SECRET        string
	CLIENT_ID            string
	RIOT_API_KEY         string
	TWITCH_CLIENT_ID     string
	TWITCH_CLIENT_SECRET string
	PATCH                string
	SEASON               string
}

type session struct {
	info           TokenResponse
	expirationDate time.Time
}

func main() {
	keys := getKeys()

	sessions := Sessions()

	oauth2 := OAuth2(AuthConfig{
		oauth_url:     "https://discord.com/api/oauth2/authorize?",
		login_url:     "/login",
		redirect_url:  "https://localhost:8000/callback",
		client_id:     keys.CLIENT_ID,
		client_secret: keys.CLIENT_SECRET,
		session_func: func(key string, value TokenResponse) {
			sessions.AddSession(key, value)
		},
	})

	oauth2.Start()
}

func getKeys() Keys {
	jsonFile, err := os.Open("keys.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	var keys Keys
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &keys)

	return keys
}
