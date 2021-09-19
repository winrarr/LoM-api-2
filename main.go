package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

func main() {
	keys := getKeys()

	oauth2 := OAuth2(AuthConfig{
		OAUTH_URL:     "https://discord.com/api/oauth2/authorize?",
		LOGIN_URL:     "/login",
		REDIRECT_URL:  "https://localhost:8000/callback",
		CLIENT_ID:     keys.CLIENT_ID,
		CLIENT_SECRET: keys.CLIENT_SECRET,
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
