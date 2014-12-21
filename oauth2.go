package main

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

var (
	client_id     string = "yp3phbkxtf8s49kxdbeqe3d2z6wjhsy6"
	client_secret string = "8TY7KjStRkXJcvPtSp7CgGn6QSWqSv3C"

	scope         string = "sc2.profile"
	redirect_uri  string = "https://localhost"
	response_type string = "code"
	state_length  int    = 32

	protocol   string = "https"
	oauth_host string = "battle.net"
	api_host   string = "api.battle.net"

	authorize_uri = "/oauth/authorize"
	token_uri     = "/oauth/token"
	api_uris      = map[string]string{
		"profile":     "/account/user/id",
		"battletag":   "/account/user/battletag",
		"sc2.profile": "/sc2/profile/user",
	}
)

func RequestPermission(region BattleNetRegion) (url string, state string) {
	conf := &oauth2.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		Scopes:       []string{"sc2.profile"},
		RedirectURL:  redirect_uri,
		Endpoint: oauth2.Endpoint{
			AuthURL:  EndpointUrl(authorize_uri, region),
			TokenURL: EndpointUrl(token_uri, region),
		},
	}

	state, _ = RandState(state_length)
	url = conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return url, state
}

func RequestToken(conf *oauth2.Config, code string) (token *oauth2.Token, err error) {
	token, err = conf.AuthenticatedExchange(oauth2.NoContext, code)

	return token, err
}

func RandState(length int) (state string, err error) {
	rb := make([]byte, length)
	_, err = rand.Read(rb)

	if err != nil {
		state = base64.URLEncoding.EncodeToString(rb)
	}

	return state, err
}

func AuthGet(url string, access_token string) *http.Response {
	client := new(http.Client)
	r, _ := http.NewRequest("GET", url, nil)

	AuthHeaders(r, access_token)
	resp, err := client.Do(r)
	// defer resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	return resp
}

func AuthHeaders(r *http.Request, access_token string) {
	r.Header.Add("Authorization", "Bearer "+access_token)
}

func ApiUri(file string, region BattleNetRegion) string {
	return protocol + "://" + region.ApiDomain() + "." + api_host + file
}

func EndpointUrl(file string, region BattleNetRegion) string {
	return protocol + "://" + region.Domain() + file
}
