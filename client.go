package main

// Client model logic. A client can be considered a User.

import (
	"github.com/Starbow/erosd/buffers"
	"log"
	"sync"
)

var _ = log.Ldate

type ClientLockoutManager struct {
	locks map[int64]*sync.RWMutex
	sync.RWMutex
}

type ClientCache struct {
	clients map[int64]*Client
	sync.RWMutex
}

var clientCache *ClientCache
var clientLockouts *ClientLockoutManager
var clientVetoes map[int64][]*Map
var clientCharacters map[int64][]*BattleNetCharacter
var clientRegionStats map[int64]map[BattleNetRegion]*ClientRegionStats

func initClientCaches() {
	clientCache = &ClientCache{
		clients: make(map[int64]*Client),
	}
	clientLockouts = &ClientLockoutManager{
		locks: make(map[int64]*sync.RWMutex),
	}
	clientVetoes = make(map[int64][]*Map)
	clientCharacters = make(map[int64][]*BattleNetCharacter)
	clientRegionStats = make(map[int64]map[BattleNetRegion]*ClientRegionStats)
}
func (cl *ClientLockoutManager) LockId(id int64) {
	cl.Lock()
	lock, ok := cl.locks[id]
	if !ok {
		var newLock sync.RWMutex
		lock = &newLock
		cl.locks[id] = lock
	}
	cl.Unlock()

	lock.Lock()
}

func (cl *ClientLockoutManager) UnlockId(id int64) {
	cl.Lock()
	lock, ok := cl.locks[id]
	if !ok {
		var newLock sync.RWMutex
		lock = &newLock
		cl.locks[id] = lock
	}
	cl.Unlock()

	lock.Unlock()
}

func (cl *ClientLockoutManager) LockIds(id ...int64) {
	for x := range id {
		cl.LockId(id[x])
	}
}
func (cl *ClientLockoutManager) UnlockIds(id ...int64) {
	for x := range id {
		cl.UnlockId(id[x])
	}
}

type Client struct {
	Id int64 `db:"id"`

	//Nickname
	Username string `db:"username"`

	//Record TrueSkill for posterity
	RatingMean   float64 `db:"rating_mean"`   // TrueSkill Mean
	RatingStdDev float64 `db:"rating_stddev"` // TrueSkill Standard Deviation

	LadderPoints int64 `db:"ladder_points"` // Global ladder points

	//Display this ranking to the world.
	LadderSearchRadius int64           `db:"ladder_search_radius"` // Search Radius.
	LadderSearchRegion BattleNetRegion `db:"ladder_search_region"`

	PendingMatchmakingId         *int64 `db:"matchmaking_pending_match_id"`
	PendingMatchmakingOpponentId *int64 `db:"matchmaking_pending_opponent_id"`
	PendingMatchmakingRegion     int64  `db:"matchmaking_pending_region"`

	Wins      int64 `db:"ladder_wins"`
	Losses    int64 `db:"ladder_losses"`
	Forfeits  int64 `db:"ladder_forefeits"`
	Walkovers int64 `db:"ladder_walkovers"`
}

type ClientRegionStats struct {
	Id       int64  `db:"id"`
	ClientId *int64 `db:"client_id"`

	Region BattleNetRegion `db:"region"`

	//Record TrueSkill for posterity
	RatingMean   float64 `db:"rating_mean"`   // TrueSkill Mean
	RatingStdDev float64 `db:"rating_stddev"` // TrueSkill Standard Deviation

	LadderPoints int64 `db:"ladder_points"` // ladder points

	Wins      int64 `db:"ladder_wins"`
	Losses    int64 `db:"ladder_losses"`
	Forfeits  int64 `db:"ladder_forefeits"`
	Walkovers int64 `db:"ladder_walkovers"`
}

func NewClient(id int64) *Client {
	client := &Client{
		Id:                 id,
		RatingMean:         25,
		RatingStdDev:       float64(25) / float64(3),
		LadderSearchRadius: 1,
		LadderPoints:       ladderStartingPoints,
	}

	return client
}

func (c *Client) Vetoes() []*Map {
	clientLockouts.LockId(c.Id)
	v, ok := clientVetoes[c.Id]
	if !ok {

		vetoRows, err := dbMap.Select(&MapVeto{}, "SELECT * FROM map_vetoes WHERE ClientId=?", c.Id)
		if err != nil {
			clientVetoes[c.Id] = make([]*Map, 0, 15)
			for x := range vetoRows {
				v := vetoRows[x].(*MapVeto)
				m, ok := maps[v.MapId]
				if ok {
					clientVetoes[c.Id] = append(clientVetoes[c.Id], m)
				}
			}
		}

	}
	clientLockouts.UnlockId(c.Id)
	return v
}

func (c *Client) Refresh() {
	clientLockouts.LockId(c.Id)
	defer clientLockouts.UnlockId(c.Id)
	err := dbMap.SelectOne(c, "SELECT * FROM clients WHERE id=? LIMIT 1", c.Id)

	if err != nil {
		log.Println("Error refreshing client", c.Id)
	}

	if _, ok := clientRegionStats[c.Id]; ok {
		for x := range clientRegionStats[c.Id] {
			clientRegionStats[c.Id][x].Refresh()
		}
	}
}

func (c *Client) IsOnline() bool {
	for _, v := range clientConnections {
		if v == nil {
			continue
		}
		if v.client.Id == c.Id {
			return true
		}
	}

	return false
}

func (c *ClientCache) Get(id int64) *Client {
	c.RLock()

	client, ok := c.clients[id]

	if !ok {
		//Concurrently acquiring the lock like this is probably terrible.
		c.RUnlock()
		c.Lock()
		defer c.Unlock()
		var newClient Client

		err := dbMap.SelectOne(&newClient, "SELECT * FROM clients WHERE id=? LIMIT 1", id)
		if err != nil || newClient.Id == 0 {

			return nil
		}

		client = &newClient
		c.clients[id] = client
	} else {
		defer c.RUnlock()
	}

	return client
}

func (c *Client) GetLadderPoints(region BattleNetRegion) int64 {
	stats, err := c.RegionStats(region)
	if err != nil {
		return 0
	} else {
		return stats.LadderPoints
	}
}

// Have Client c defeat Client o and update their ratings.
func (client *Client) Defeat(opponent *Client, region BattleNetRegion) float64 {

	// Update W/L
	client.Wins += 1
	opponent.Losses += 1

	var quality float64
	client.RatingMean, client.RatingStdDev, opponent.RatingMean, opponent.RatingStdDev, quality = calculateNewRating(client.Id, opponent.Id, client.RatingMean, client.RatingStdDev, opponent.RatingMean, opponent.RatingStdDev)
	client.LadderPoints, opponent.LadderPoints = calculateNewPoints(client.LadderPoints, opponent.LadderPoints)

	regionStats, err := client.RegionStats(region)

	if err != nil {
		return quality
	}
	opponentRegionStats, err := opponent.RegionStats(region)
	if err != nil {
		return quality
	}

	regionStats.RatingMean, regionStats.RatingStdDev, opponentRegionStats.RatingMean, opponentRegionStats.RatingStdDev, quality = calculateNewRating(client.Id, opponent.Id, regionStats.RatingMean, regionStats.RatingStdDev, opponentRegionStats.RatingMean, opponentRegionStats.RatingStdDev)
	regionStats.LadderPoints, opponentRegionStats.LadderPoints = calculateNewPoints(regionStats.LadderPoints, opponentRegionStats.LadderPoints)
	regionStats.Wins += 1
	opponentRegionStats.Losses += 1

	dbMap.Update(regionStats, opponentRegionStats)

	log.Println(client.Username, client.LadderPoints, "defeated", opponent.Username, opponent.LadderPoints)

	return quality
}

func (c *Client) ForfeitMatchmadeMatch() {
	if c.PendingMatchmakingId != nil {
		var opponent *Client = nil
		if c.PendingMatchmakingOpponentId != nil {
			opponent = clientCache.Get(*c.PendingMatchmakingOpponentId)
		}
		if opponent.PendingMatchmakingId != nil && *opponent.PendingMatchmakingId == *c.PendingMatchmakingId {
			opponent.Defeat(c, BattleNetRegion(c.PendingMatchmakingRegion))
			c.Forfeits += 1
			opponent.Walkovers += 1
			matchmaker.EndMatch(*c.PendingMatchmakingId)
			dbMap.Update(c, opponent)
			log.Println(c.Username, "forfeited")

			go func() {
				c.BroadcastStatsMessage()
				c.BroadcastMatchmakingIdle()
			}()
			go func() {
				opponent.BroadcastStatsMessage()
				opponent.BroadcastMatchmakingIdle()
			}()
		} else {
			matchmaker.EndMatch(*c.PendingMatchmakingId)
		}
	}
}

// Check if the client is matched against the opponent in matchmaking.
func (c *Client) IsMatchedWith(opponent *Client) bool {
	if c.PendingMatchmakingId == nil {
		return true
	}

	return (*c.PendingMatchmakingOpponentId == opponent.Id)
}

// Check if the client is matched against the opponent in matchmaking.
func (c *Client) IsOnMap(id int64) bool {
	if c.PendingMatchmakingId == nil {
		return true
	}

	match := matchmaker.Match(*c.PendingMatchmakingId)
	if match == nil {
		return true
	}

	return (match.MapId != nil && *match.MapId == id)
}

// Generate a UserStats protocol buffer message from this client.
func (c *Client) UserStatsMessage() *protobufs.UserStats {
	var user protobufs.UserStats
	user.Points = &c.LadderPoints
	user.Username = &c.Username
	user.SearchRadius = &c.LadderSearchRadius
	user.Wins = &c.Wins
	user.Losses = &c.Losses
	user.Walkovers = &c.Walkovers
	user.Forfeits = &c.Forfeits
	user.Region = make([]*protobufs.UserRegionStats, 0, len(ladderActiveRegions))

	for _, region := range ladderActiveRegions {
		stats, err := c.RegionStats(region)
		if stats != nil && err == nil {
			user.Region = append(user.Region, stats.UserRegionStatsMessage())
		} else {
			log.Println(region, err, stats)
		}
	}

	return &user
}

// Broadcast a stats message to this client if they are connected.
func (c *Client) BroadcastStatsMessage() {
	message := c.UserStatsMessage()
	c.Broadcast("USU", message)
}

func (c *Client) BroadcastMatchmakingIdle() {
	c.Broadcast("MMI", nil)
}

// Check if the client can queue in this region.
func (c *Client) HasRegion(region BattleNetRegion) bool {
	count, _ := dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE ClientId=? and Region=? and IsVerified=?", c.Id, region, true)

	return count > 0
}

func (c *Client) Characters() (characters []*BattleNetCharacter, err error) {
	clientLockouts.LockId(c.Id)
	defer clientLockouts.UnlockId(c.Id)

	var ok bool = false
	if characters, ok = clientCharacters[c.Id]; ok {
		return
	}

	chars, err := dbMap.Select(&BattleNetCharacter{}, "SELECT * FROM battle_net_characters WHERE ClientId=?", c.Id)
	if err != nil {
		characters = nil
		return
	}

	characters = make([]*BattleNetCharacter, len(chars))

	clientCharacters[c.Id] = characters

	characterCache.Lock()
	for x := range chars {
		character := chars[x].(*BattleNetCharacter)
		characters[x] = character

		characterCache.characterIds[character.Id] = character
		characterCache.profileIds[character.ProfileIdString()] = character
	}

	characterCache.Unlock()

	return
}

func (c *Client) RegionStats(region BattleNetRegion) (regionStats *ClientRegionStats, err error) {
	clientLockouts.LockId(c.Id)
	defer clientLockouts.UnlockId(c.Id)

	var ok bool = false
	if _, ok = clientRegionStats[c.Id]; !ok {
		clientRegionStats[c.Id] = make(map[BattleNetRegion]*ClientRegionStats)
	}

	if regionStats, ok = clientRegionStats[c.Id][region]; ok {
		return
	}

	var stats ClientRegionStats

	err = dbMap.SelectOne(&stats, "SELECT * FROM client_region_stats WHERE client_id=? and region=?", c.Id, int64(region))
	if err != nil || stats.Id == 0 {
		stats.ClientId = &c.Id
		stats.Forfeits = 0
		stats.Losses = 0
		stats.LadderPoints = ladderStartingPoints
		stats.RatingMean = 25
		stats.RatingStdDev = float64(24) / float64(3)
		stats.Walkovers = 0
		stats.Wins = 0
		stats.Region = region

		err = dbMap.Insert(&stats)
	}

	clientRegionStats[c.Id][region] = &stats
	return &stats, nil
}

func (crs *ClientRegionStats) Refresh() {
	err := dbMap.SelectOne(crs, "SELECT * FROM client_region_stats WHERE id=? LIMIT 1", crs.Id)
	if err != nil {
		log.Println("Error refreshing CRS", crs.ClientId, crs.Id)
	}
}

func (crs *ClientRegionStats) UserRegionStatsMessage() *protobufs.UserRegionStats {
	var stats protobufs.UserRegionStats
	var region protobufs.Region = protobufs.Region(crs.Region)
	stats.Points = &crs.LadderPoints
	stats.Wins = &crs.Wins
	stats.Losses = &crs.Losses
	stats.Walkovers = &crs.Walkovers
	stats.Forfeits = &crs.Forfeits
	stats.Region = &region

	return &stats
}

func (client *Client) SendBroadcastAlert(predefined int32, message string) {
	var bufmsg protobufs.BroadcastAlert
	bufmsg.Message = &message
	bufmsg.Predefined = &predefined

	client.Broadcast("ALT", &bufmsg)
}
