package main

import (
	//"log"
	"encoding/json"
	"errors"
	"fmt"
	protobufs "github.com/Starbow/erosd/buffers"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type BattleNetCharacterCache struct {
	characterIds map[int64]*BattleNetCharacter
	profileIds   map[string]*BattleNetCharacter
	sync.RWMutex
}

var (
	battleNetProfileRegex = regexp.MustCompile(`http://((www\.battlenet\.com\.(cn))|((\w+)\.battle\.net))/sc2/\w+/profile/(\d+)/(\d+)/(.+)/$`)
	ErrBadRegion          = errors.New("A bad region code was passed.")
	ErrBattleNetLookup    = errors.New("An error occured while Battle.Net lookupinginging")
	characterCache        *BattleNetCharacterCache
)

func initBattleNet() {
	characterCache = &BattleNetCharacterCache{
		characterIds: make(map[int64]*BattleNetCharacter),
		profileIds:   make(map[string]*BattleNetCharacter),
	}
}

type BattleNetRegion int

const (
	BATTLENET_REGION_UNKNOWN BattleNetRegion = 0
	BATTLENET_REGION_NA      BattleNetRegion = 1
	BATTLENET_REGION_EU      BattleNetRegion = 2
	BATTLENET_REGION_KR      BattleNetRegion = 3
	BATTLENET_REGION_CN      BattleNetRegion = 5
	BATTLENET_REGION_SEA     BattleNetRegion = 6
)

func (c *BattleNetCharacterCache) GetId(id int64) (character *BattleNetCharacter) {
	c.RLock()

	character, ok := c.characterIds[id]

	if !ok {
		//Concurrently acquiring the lock like this is probably terrible.

		c.RUnlock()
		c.Lock()
		var newChar BattleNetCharacter

		err := dbMap.SelectOne(&newChar, "SELECT * FROM battle_net_characters WHERE Id=? LIMIT 1", id)
		if err != nil || newChar.Id == 0 {
			return nil
		}

		character = &newChar
		c.characterIds[id] = character
		c.profileIds[character.ProfileIdString()] = character

		defer c.Unlock()
	} else {
		defer c.RUnlock()
	}

	return character
}

func (c *BattleNetCharacterCache) Get(region BattleNetRegion, subregion, profileId int) (character *BattleNetCharacter) {
	c.RLock()

	character, ok := c.profileIds[BattleNetProfileIdString(region, subregion, profileId)]

	if !ok {
		//Concurrently acquiring the lock like this is probably terrible.
		c.RUnlock()
		c.Lock()
		var newChar BattleNetCharacter

		err := dbMap.SelectOne(&newChar, "SELECT * FROM battle_net_characters WHERE Region=? and SubRegion=? and ProfileId=? and IsVerified=? LIMIT 1", region, subregion, profileId, true)
		if err != nil || newChar.Id == 0 {
			log.Println(err)
			return nil
		}

		character = &newChar
		c.characterIds[character.Id] = character
		c.profileIds[character.ProfileIdString()] = character

		defer c.Unlock()
	} else {
		defer c.RUnlock()
	}

	return character
}

func (region BattleNetRegion) Language() string {
	switch region {
	case BATTLENET_REGION_NA, BATTLENET_REGION_EU, BATTLENET_REGION_SEA:
		return "en"
	case BATTLENET_REGION_KR:
		return "ko"
	case BATTLENET_REGION_CN:
		return "zh"
	}

	return ""
}

// Matchmaking region is different to BattleNet region

func (region BattleNetRegion) MatchmakingRegion() int64 {
	switch region {
	case BATTLENET_REGION_NA:
		return MATCHMAKING_REGION_NA
	case BATTLENET_REGION_EU:
		return MATCHMAKING_REGION_EU
	case BATTLENET_REGION_KR:
		return MATCHMAKING_REGION_KR
	case BATTLENET_REGION_SEA:
		return MATCHMAKING_REGION_SEA
	case BATTLENET_REGION_CN:
		return MATCHMAKING_REGION_CN
	}

	return 0
}

func (region BattleNetRegion) Domain() string {

	switch region {
	case BATTLENET_REGION_NA:
		return "us.battle.net"
	case BATTLENET_REGION_EU:
		return "eu.battle.net"
	case BATTLENET_REGION_KR:
		return "kr.battle.net"
	case BATTLENET_REGION_SEA:
		return "sea.battle.net"
	case BATTLENET_REGION_CN:
		return "www.battlenet.com.cn"
	}

	return ""
}

type BattleNetAPIProfile struct {
	Code     int64 `json:"code"`
	Portrait struct {
		Offset int    `json:"offset"`
		Url    string `json:"url"`
	} `json:"portrait"`
}

type BattleNetCharacter struct {
	Id                            int64
	ClientId                      int64
	AddTime                       int64
	Region                        BattleNetRegion
	SubRegion                     int
	ProfileId                     int
	CharacterName                 string
	CharacterCode                 int
	InGameProfileLink             string
	IsVerified                    bool
	VerificationRequestedPortrait int
}

func (c *BattleNetCharacter) ApiUrl() string {
	domain := c.Region.Domain()
	if domain == "" {
		return ""
	}

	return fmt.Sprintf("http://%s/api/sc2/profile/%d/%d/%s/", domain, c.ProfileId, c.SubRegion, c.CharacterName)
}

func (c *BattleNetCharacter) Url() string {
	domain := c.Region.Domain()
	language := c.Region.Language()
	if domain == "" {
		return ""
	}

	return fmt.Sprintf("http://%s/sc2/%s/profile/%d/%d/%s/", domain, language, c.ProfileId, c.SubRegion, c.CharacterName)
}

// Get the current portrait offset.
func (c *BattleNetCharacter) GetPortrait() (portrait int, err error) {
	url := c.ApiUrl()
	if url == "" {
		return -1, ErrBadRegion
	}

	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var profile BattleNetAPIProfile
	err = decoder.Decode(&profile)
	if err != nil {
		return -1, err
	}

	if profile.Code > 0 {
		return -1, ErrBattleNetLookup
	}

	if profile.Portrait.Url == "" {
		return -1, ErrBattleNetLookup
	}

	return profile.Portrait.Offset, nil
}

//Set the requested portrait to a value between 0-2 that isn't the current one
func (c *BattleNetCharacter) SetVerificationPortrait() error {
	current, err := c.GetPortrait()
	if err != nil {
		return err
	}

	verification := current
	for {
		verification = rand.Intn(3)

		if verification != current {
			break
		}
	}

	c.VerificationRequestedPortrait = verification
	return nil
}

//Compare the requested portrait to the current portrait.
func (c *BattleNetCharacter) CheckVerificationPortrait() (bool, error) {
	current, err := c.GetPortrait()
	if err != nil {
		return false, err
	}

	return (c.VerificationRequestedPortrait == current), nil
}

func (c *BattleNetCharacter) CharacterMessage() *protobufs.Character {
	var (
		character protobufs.Character
		region    protobufs.Region = protobufs.Region(c.Region)
		subregion int32            = int32(c.SubRegion)
		profileid int32            = int32(c.ProfileId)
		portrait  int32            = int32(c.VerificationRequestedPortrait)
		code      int32            = int32(c.CharacterCode)
		link      string           = c.Url()
	)

	character.Region = &region
	character.Subregion = &subregion
	character.ProfileId = &profileid
	character.CharacterName = &c.CharacterName
	character.CharacterCode = &code
	character.IngameProfileLink = &c.InGameProfileLink
	character.ProfileLink = &link
	character.Verified = &c.IsVerified
	character.VerificationPortrait = &portrait

	return &character

}

func (c *BattleNetCharacter) ProfileIdString() string {
	return BattleNetProfileIdString(c.Region, c.SubRegion, c.ProfileId)
}

func BattleNetProfileIdString(region BattleNetRegion, subregion, profileId int) string {
	return fmt.Sprintf("%d-S2-%d-%d", region, subregion, profileId)
}

func NewBattleNetCharacter(region BattleNetRegion, subregion, profileId int, name string) (character *BattleNetCharacter) {
	return &BattleNetCharacter{
		Region:        region,
		SubRegion:     subregion,
		ProfileId:     profileId,
		CharacterName: name,
		IsVerified:    false,
		AddTime:       time.Now().Unix(),
	}
}

func ParseBattleNetRegion(region string) (code BattleNetRegion) {
	switch region {
	case "US", "us", "NA", "na":
		return BATTLENET_REGION_NA
	case "EU", "eu":
		return BATTLENET_REGION_EU
	case "KR", "kr":
		return BATTLENET_REGION_KR
	case "SEA", "sea":
		return BATTLENET_REGION_SEA
	case "CN", "cn", "TW", "tw":
		return BATTLENET_REGION_CN
	}

	return BATTLENET_REGION_UNKNOWN
}

func ParseBattleNetProfileUrl(url string) (region BattleNetRegion, subregion, id int, name string) {
	res := battleNetProfileRegex.FindAllStringSubmatch(url, -1)
	if len(res) == 0 {
		return
	}

	region = ParseBattleNetRegion(res[0][5])
	subregion, _ = strconv.Atoi(res[0][7])
	id, _ = strconv.Atoi(res[0][6])
	name = res[0][8]

	return
}
