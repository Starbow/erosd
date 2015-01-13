package main

import (
	"encoding/json"
	"errors"
	// "golang.org/x/oauth2"
	"github.com/Sikian/oauth2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

var (
	oauthClientSecret string
	oauthClientId     string

	scope            string = "sc2.profile"
	oauthCodeTimeout int64
	oauthRedirectUri string
	response_type    string = "code"
	state_length     int    = 64

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

type OAuthRequest struct {
	state  string
	code   string
	token  *oauth2.Token
	region BattleNetRegion
	config *oauth2.Config
	conn   *ClientConnection
}

type BnetInfo struct {
	AccountId  int       `json:"id"`
	Battletag  string    `json:"battletag"`
	Characters []Sc2Char `json:"characters"`
}

type Sc2Char struct {
	ProfileId   int    `json:"id"`
	Realm       int    `json:"realm"`
	DisplayName string `json:"displayname"`
	ClanTag     string `json:"clantag"`
	Portrait    struct {
		// Offset int    `json:"offset"`
		X   int    `json:"x"`
		Y   int    `json:"y"`
		W   int    `json:"w"`
		H   int    `json:"h"`
		Url string `json:"url"`
	} `json:"portrait"`
	Career struct {
		PrimaryRace string `json:"primaryrace"`
	} `json:"career"`
}

func (oar *OAuthRequest) RequestPermission() (url string, state string) {
	oar.config = &oauth2.Config{
		ClientID:     oauthClientId,
		ClientSecret: oauthClientSecret,
		Scopes:       []string{"sc2.profile"},
		RedirectURL:  oauthRedirectUri,
		Endpoint: oauth2.Endpoint{
			AuthURL:  EndpointUrl(authorize_uri, oar.region),
			TokenURL: EndpointUrl(token_uri, oar.region),
		},
	}
	oar.state = RandState(state_length)
	url = oar.config.AuthCodeURL(oar.state, oauth2.AccessTypeOffline)

	return url, oar.state
}

func (oar *OAuthRequest) RequestToken() (token *oauth2.Token, err error) {
	oar.token, err = oar.config.AuthenticatedExchange(oauth2.NoContext, oar.code)

	return oar.token, err
}

func RandState(length int) (state string) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"

	buf := make([]byte, length)
	for j := 0; j < length; j++ {
		buf[j] = chars[rand.Intn(len(chars))]
	}

	return string(buf)
}

func (oar *OAuthRequest) AuthGet(url string, access_token string) *http.Response {
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

// Proxu for profile
func (oar *OAuthRequest) GetProfile() (bnetinfo BnetInfo, err error) {
	if oar.token == nil {
		err = errors.New("Don't have a token.")
	} else {
		resp := oar.AuthGet(ApiUri(api_uris["profile"], oar.region), oar.token.AccessToken)
		// io.Copy(os.Stdout, resp.Body)
		defer resp.Body.Close()

		b, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(b, &bnetinfo)
	}

	return bnetinfo, err
}

// Proxy for sc2 profile
func (oar *OAuthRequest) GetSC2Profile() (sc2char Sc2Char, err error) {
	var bnetinfo BnetInfo
	if oar.token == nil {
		err = errors.New("Don't have a token.")
	} else {
		resp := oar.AuthGet(ApiUri(api_uris["sc2.profile"], oar.region), oar.token.AccessToken)
		// io.Copy(os.Stdout, resp.Body)
		defer resp.Body.Close()

		b, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(b, &bnetinfo)

		if err == nil {
			if len(bnetinfo.Characters) == 0 {
				err = errors.New("No characters found for this region.")
			} else {
				sc2char = bnetinfo.Characters[0]
			}
		}
	}

	return sc2char, err
}

func AuthHeaders(r *http.Request, access_token string) {
	r.Header.Add("Authorization", "Bearer "+access_token)
}

func ApiUri(file string, region BattleNetRegion) string {
	return protocol + "://" + region.ApiDomain() + file
}

func EndpointUrl(file string, region BattleNetRegion) string {
	return protocol + "://" + region.Domain() + file
}

/*
 * Eros helper
 *
 */

// Requests token and fetches the SC2 profile
func (oar *OAuthRequest) getCharInfo(code string) (char Sc2Char, proto_char *BattleNetCharacter, err error) {
	oar.code = code

	oar.conn.logger.Println("Requesting OAuth token.")
	oar.RequestToken()

	oar.conn.logger.Println("Adding new battlenet character.")
	char, proto_char, err = AddOAuthProfile(oar)

	if err != nil {
		return
	} else {
		delete(activeOAuths, oar.state)
	}
	return
}

// Gets the SC2 profile for an authorized request
func AddOAuthProfile(oar *OAuthRequest) (profile Sc2Char, character *BattleNetCharacter, err error) {
	profile, err = oar.GetSC2Profile()

	if err != nil {
		// oar.conn.logger.Println(err)
		return
	}

	region := oar.region
	subregion := profile.Realm
	id := profile.ProfileId
	name := profile.DisplayName

	count, err := dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE Region=? and SubRegion=? and ProfileId=?", region, subregion, id)

	if err != nil {
		err = ErosErrors(101)
		return
	}

	if count > 0 {
		// Check if there's any profiles which have been disabled for this user
		count, err = dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE Region=? and SubRegion=? and ProfileId=? and ClientId=? and Enabled=?", region, subregion, id, oar.conn.client.Id, false)

		if err != nil {
			oar.conn.logger.Println(err)
			err = ErosErrors(101)
			return
		}

		if count == 0 {
			// Profile exists but is not from that user
			err = ErosErrors(202)
			return
		}
	}

	character = NewBattleNetCharacter(region, subregion, id, name)
	character.ClientId = &oar.conn.client.Id
	character.IsVerified = true
	character.Enabled = true

	if err != nil {
		oar.conn.logger.Println(err)
		err = ErosErrors(203)
		return
	}

	if count == 0 {
		// Insert the character if it's a new one
		oar.conn.logger.Println("Inserting new character.")
		err = dbMap.Insert(character)
	} else {
		// count, err = dbMap.Update(character)
		oar.conn.logger.Println("Reenabling character.")
		_, err = dbMap.Exec("UPDATE battle_net_characters SET Enabled=? WHERE Region=? and SubRegion=? and ProfileId=?", true, region, subregion, id)

	}

	if err != nil {
		oar.conn.logger.Println("Error inserting character", err)
		err = ErosErrors(102)
		return
	}

	// This should be its own function
	characterCache.Lock()
	characterCache.characterIds[character.Id] = character
	characterCache.profileIds[character.ProfileIdString()] = character
	characterCache.Unlock()

	delete(clientCharacters, oar.conn.client.Id)

	return
}
