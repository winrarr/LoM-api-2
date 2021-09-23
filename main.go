package main

import (
	"LoM-api/oauth2"
	"LoM-api/sessions"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	keys := getKeys()

	sessions := sessions.Sessions()

	oauth2 := oauth2.OAuth2(oauth2.AuthConfig{
		Oauth_url:     "https://discord.com/api/oauth2/authorize?",
		Login_url:     "/login",
		Redirect_uri:  "https://localhost:8000/callback",
		Client_id:     keys.CLIENT_ID,
		Client_secret: keys.CLIENT_SECRET,
		Scope:         "identify",
		Session_func: func(key string, value oauth2.TokenResponse) {
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
