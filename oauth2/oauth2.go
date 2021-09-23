package oauth2

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
	Oauth_url     string
	Login_url     string
	Redirect_uri  string
	Client_id     string
	Client_secret string
	Scope         string
	Session_func  func(string, TokenResponse)
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

	r.HandleFunc(auth.config.Login_url, auth.Login)
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
		auth.config.Oauth_url+
			"client_id="+auth.config.Client_id+
			"&redirect_uri="+auth.config.Redirect_uri+
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
	println(code, state.Value, session.Value)

	info, err := auth.getAccessToken(code)
	if err != nil {
		log.Println(err)
		return
	}
	auth.config.Session_func(session.Value, *info)
}

type tokenRequest struct {
	Client_id     string `json:"client_id"`
	Client_secret string `json:"client_secret"`
	Redirect_uri  string `json:"redirect_uri"`
	Grant_type    string `json:"grant_type"`
	Code          string `json:"code"`
	Scope         string `json:"scope"`
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
		Client_id:     auth.config.Client_id,
		Client_secret: auth.config.Client_secret,
		Redirect_uri:  auth.config.Redirect_uri,
		Grant_type:    "authorization_code",
		Code:          code,
		Scope:         auth.config.Scope,
	}

	var resBody TokenResponse
	postRequest("https://discord.com/api/oauth2/token", body, &resBody)
	return &resBody, nil
}

type refreshTokenRequest struct {
	Client_id     string `json:"client_id"`
	Client_secret string `json:"client_secret"`
	Grant_type    string `json:"grant_type"`
	Refresh_token string `json:"refresh_token"`
}

func (auth *auth) refreshToken(refreshToken string) (*TokenResponse, error) {
	body := refreshTokenRequest{
		Client_id:     auth.config.Client_id,
		Client_secret: auth.config.Client_secret,
		Grant_type:    "refresh_token",
		Refresh_token: refreshToken,
	}

	var resBody TokenResponse
	postRequest("https://discord.com/api/oauth2/token", body, &resBody)
	return &resBody, nil
}

func postRequest(url string, body interface{}, response interface{}) error {
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
