package main

// Client model logic. A client can be considered a User.

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/ChrisHines/GoSkills/skills"
	"github.com/ChrisHines/GoSkills/skills/trueskill"
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

func initClientCaches() {
	clientCache = &ClientCache{
		clients: make(map[int64]*Client),
	}
	clientLockouts = &ClientLockoutManager{
		locks: make(map[int64]*sync.RWMutex),
	}
	clientVetoes = make(map[int64][]*Map)
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
	Id      int64 `db:"id"`
	AuthKey string

	//Nickname
	Username string `db:"username"`

	//Record TrueSkill for posterity
	RatingMean   float64 `db:"rating_mean"`   // TrueSkill Mean
	RatingStdDev float64 `db:"rating_stddev"` // TrueSkill Standard Deviation

	//Display this ranking to the world.
	LadderPoints       int64           `db:"ladder_points"`        // iCCup ladder points
	LadderSearchRadius int64           `db:"ladder_search_radius"` // Search Radius.
	LadderSearchRegion BattleNetRegion `db:"ladder_search_region"`
	TotalQueueTime     float64         `db:"ladder_total_queue_time"`

	PendingMatchmakingId         int64 `db:"matchmaking_pending_match_id"`
	PendingMatchmakingOpponentId int64 `db:"matchmaking_pending_opponent_id"`

	Wins      int64 `db:"ladder_wins"`
	Losses    int64 `db:"ladder_losses"`
	Forefeits int64 `db:"ladder_forefeits"`
	Walkovers int64 `db:"ladder_walkovers"`
}

func NewClient() *Client {
	var client Client
	client.setRandomAuthKey()
	client.LadderPoints = ladderStartingPoints
	client.LadderSearchRadius = 1
	client.RatingMean = 25
	client.RatingStdDev = float64(25) / float64(3)

	return &client
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

// Set the client's auth key to something random
func (c *Client) setRandomAuthKey() {
	rnd := make([]byte, 64)
	rand.Read(rnd)

	c.AuthKey = base64.StdEncoding.EncodeToString(rnd)[:64]
}

// Have Client c defeat Client o and update their ratings.
func (c *Client) Defeat(o *Client) float64 {

	// Calculate the TrueSkill
	player1 := skills.NewPlayer(c.Id)
	player2 := skills.NewPlayer(o.Id)

	team1 := skills.NewTeam()
	team2 := skills.NewTeam()

	team1.AddPlayer(*player1, skills.NewRating(c.RatingMean, c.RatingStdDev))
	team2.AddPlayer(*player2, skills.NewRating(o.RatingMean, o.RatingStdDev))

	teams := []skills.Team{team1, team2}

	var calc trueskill.TwoPlayerCalc
	ratings := calc.CalcNewRatings(skills.DefaultGameInfo, teams, 1, 2)
	quality := calc.CalcMatchQual(skills.DefaultGameInfo, teams)

	c.RatingMean = ratings[*player1].Mean()
	c.RatingStdDev = ratings[*player1].Stddev()

	o.RatingMean = ratings[*player2].Mean()
	o.RatingStdDev = ratings[*player2].Stddev()

	// Update W/L
	c.Wins += 1
	o.Losses += 1

	// Update points
	// GetDifference(2000, 1000) would return -1
	// GetDifference(2000, 3000) would return 1
	difference := divisions.GetDifference(c.LadderPoints, o.LadderPoints)
	increase := ladderWinPointsBase + int64((ladderWinPointsIncrement * float64(difference)))
	decrease := ladderLosePointsBase - (int64((ladderLosePointsIncrement * float64(difference))) * -1)
	if increase < 0 {
		increase = 10
	}

	if decrease < 0 {
		decrease = 0
	}

	c.LadderPoints += increase
	o.LadderPoints -= decrease

	if c.LadderPoints < 0 {
		c.LadderPoints = 0
	}

	if o.LadderPoints < 0 {
		o.LadderPoints = 0
	}

	return quality
}

func (c *Client) ForefeitMatchmadeMatch() {
	if c.PendingMatchmakingId > 0 {
		opponent := clientCache.Get(c.PendingMatchmakingOpponentId)
		opponent.Defeat(c)
		c.Forefeits += 1
		opponent.Walkovers += 1
		matchmaker.EndMatch(c.PendingMatchmakingId, c, opponent)
		dbMap.Update(c, opponent)
	}
}

// Check if the client is matched against the opponent in matchmaking.
func (c *Client) IsMatchedWith(opponent *Client) bool {
	if c.PendingMatchmakingId == 0 {
		return true
	}

	return (c.PendingMatchmakingOpponentId == opponent.Id)
}

// Check if the client is matched against the opponent in matchmaking.
func (c *Client) IsOnMap(id int64) bool {
	if c.PendingMatchmakingId == 0 {
		return true
	}

	match := matchmaker.Match(c.PendingMatchmakingId)
	if match == nil {
		return true
	}

	return (match.MapId == id)
}

// Generate a UserStats protocol buffer message from this client.
func (c *Client) UserStatsMessage() protobufs.UserStats {
	var user protobufs.UserStats
	user.Points = &c.LadderPoints
	user.Username = &c.Username
	user.SearchRadius = &c.LadderSearchRadius
	user.Wins = &c.Wins
	user.Losses = &c.Losses

	return user
}

// Broadcast a stats message to this client if they are connected.
func (c *Client) BroadcastStatsMessage() {
	message := c.UserStatsMessage()
	c.Broadcast("USU", &message)
}

// Check if the client can queue in this region.
func (c *Client) HasRegion(region BattleNetRegion) bool {
	count, _ := dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE ClientId=? and Region=?", c.Id, region)

	return count > 0
}
