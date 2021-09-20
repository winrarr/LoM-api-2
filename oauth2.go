package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type auth struct {
	config AuthConfig
}

type AuthConfig struct {
	oauth_url     string
	login_url     string
	redirect_url  string
	client_id     string
	client_secret string
	session_func  func(string, TokenResponse)
}

func OAuth2(config AuthConfig) auth {
	return auth{
		config: config,
	}
}

func (auth *auth) Config(config AuthConfig) {
	auth.config = config
}

func (auth *auth) Start() {
	r := mux.NewRouter()

	r.HandleFunc(auth.config.login_url, auth.Login)
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
		auth.config.oauth_url+
			"client_id="+auth.config.client_id+
			"&redirect_uri="+auth.config.redirect_url+
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
	auth.config.session_func(session.Value, *info)
}

type tokenRequest struct {
	Client_id     string
	Client_secret string
	Redirect_url  string
	Grant_type    string
	Code          string
	Scope         string
}

type TokenResponse struct {
	Access_token  string
	Expires_in    string
	Refresh_token string
	Scope         string
	Token_type    string
}

func (auth *auth) getAccessToken(code string) (*TokenResponse, error) {
	body := tokenRequest{
		Client_id:     auth.config.client_id,
		Client_secret: auth.config.client_secret,
		Redirect_url:  auth.config.redirect_url,
		Grant_type:    "authorization_code",
		Code:          code,
		Scope:         "identify",
	}

	var resBody TokenResponse
	auth.postRequest("https://discord.com/api/oauth2/token", body, &resBody)
	return &resBody, nil
}

type refreshTokenRequest struct {
	Client_id     string
	Client_secret string
	Grant_type    string
	Refresh_token string
}

func (auth *auth) refreshToken(refreshToken string) (*TokenResponse, error) {
	body := refreshTokenRequest{
		Client_id:     auth.config.client_id,
		Client_secret: auth.config.client_secret,
		Grant_type:    "refresh_token",
		Refresh_token: refreshToken,
	}

	var resBody TokenResponse
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
