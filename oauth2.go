package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type auth struct {
	sessions map[string]sessionInfo
	config   AuthConfig
}

type sessionInfo struct {
	info           tokenResponse
	expirationDate time.Time
}

type AuthConfig struct {
	OAUTH_URL     string
	LOGIN_URL     string
	REDIRECT_URL  string
	CLIENT_ID     string
	CLIENT_SECRET string
}

func OAuth2(config AuthConfig) auth {
	return auth{
		sessions: make(map[string]sessionInfo),
		config:   config,
	}
}

func (auth *auth) Config(config AuthConfig) {
	auth.config = config
}

func (auth *auth) Start() {
	r := mux.NewRouter()

	r.HandleFunc(auth.config.LOGIN_URL, auth.Login)
	r.Path("/callback").
		Queries("code", "", "state", "").
		HandlerFunc(auth.callback).
		Methods("GET")

	http.ListenAndServe(":8000", r)
}

func (auth *auth) Login(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		// Do not login
		log.Println(err)
		return
	}

	http.Redirect(w, r,
		auth.config.OAUTH_URL+
			"client_id="+auth.config.CLIENT_ID+
			"&redirect_uri="+auth.config.REDIRECT_URL+
			"&response_type=code"+
			"&scope=identify"+
			"&state="+state.Value+
			"&prompt=consent", http.StatusTemporaryRedirect)
}

func (auth *auth) callback(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	code := r.FormValue("code")

	stateRes := r.FormValue("state")
	state, err := r.Cookie("state")
	if err != nil {
		// Do not login
		log.Println(err)
		return
	}

	if stateRes != state.Value {
		// Do not login
		return
	}

	session, err := r.Cookie("session")
	if err != nil {
		// Do not login
		log.Println(err)
		return
	}
	println(code, state, session.Value)

	info, err := auth.getAccessToken(code)
	if err != nil {
		log.Println(err)
		return
	}
	auth.sessions[session.Value] = sessionInfo{*info, time.Now()}
}

type tokenRequest struct {
	Client_id     string
	Client_secret string
	Redirect_url  string
	Grant_type    string
	Code          string
	Scope         string
}

type tokenResponse struct {
	Access_token  string
	Expires_in    string
	Refresh_token string
	Scope         string
	Token_type    string
}

func (auth *auth) getAccessToken(code string) (*tokenResponse, error) {
	body := tokenRequest{
		Client_id:     auth.config.CLIENT_ID,
		Client_secret: auth.config.CLIENT_SECRET,
		Redirect_url:  auth.config.REDIRECT_URL,
		Grant_type:    "authorization_code",
		Code:          code,
		Scope:         "identify",
	}

	var resBody tokenResponse
	auth.postRequest("https://discord.com/api/oauth2/token", body, &resBody)
	return &resBody, nil
}

type refreshTokenRequest struct {
	Client_id     string
	Client_secret string
	Grant_type    string
	Refresh_token string
}

func (auth *auth) refreshToken(refreshToken string) (*tokenResponse, error) {
	body := refreshTokenRequest{
		Client_id:     auth.config.CLIENT_ID,
		Client_secret: auth.config.CLIENT_SECRET,
		Grant_type:    "refresh_token",
		Refresh_token: refreshToken,
	}

	var resBody tokenResponse
	auth.postRequest("https://discord.com/api/oauth2/token", body, &resBody)
	return &resBody, nil
}

func (auth *auth) postRequest(url string, body interface{}, response interface{}) error {
	jsonBytes, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
