package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"os"
	"testing"
)

var region = BATTLENET_REGION_NA

func print(s string){
	fmt.Fprintln(os.Stderr, "[iccup] %v", s)
}

func init() {
	print("init()")
	dbType = "sqlite3"
	dbConnectionString = "erosd.sqlite3"
	initDb()
	initChat()
	initClientCaches()

	ladder.InitDivisions()

	testMode=true;
	ladderStartingPoints = 1200;
}

func createMockUser(t *testing.T, id int64) *ClientConnection{
	var user RealUser
	user.Username = fmt.Sprintf("MockUser%d", id)
	user.AuthToken = fmt.Sprintf("%d", id)
	user.Email = fmt.Sprintf("mockuser%d@starbowmod.com", id)
	user.IsActive = true
	err := dbMap.Insert(&user)
	if err != nil {
		fmt.Println(err)
	}

	// c:= NewSimulatedUser(1, 2000, 3)
	site_user := GetRealUser(user.Username, user.AuthToken)

	client := clientCache.Get(site_user.Id)
	if client == nil {
		client = NewClient(site_user.Id)
		err := dbMap.Insert(client)
		clientCache.clients[client.Id] = client
		client.Username = site_user.Username

		dbMap.Update(client)
		if err != nil {
			fmt.Println(err)
		}

	}

	// fmt.Println("Creating new mock user", site_user);
	
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	write := &bytes.Buffer{}
	c := &ClientConnection{
		id:            id,
		authenticated: false,
		chatRooms:     make(map[string]*ChatRoom),
		conn:          conn,
		connType:      CLIENT_CONNECTION_TYPE_SOCKET,
		client:        client,
		logger:        log.New(os.Stdout, fmt.Sprintf("chat-%d:", id), log.Ldate|log.Ltime|log.Lshortfile),
		writer:        bufio.NewWriter(write),
	}

	updateClientDivisions(t, c.client)

	return c;
}

func updateClientDivisions(t *testing.T, c *Client){
	c.Division, c.DivisionRank =  divisions.GetDivision(c.LadderPoints)

	regionStats, err := c.RegionStats(region)
	
	if(err != nil){t.Fatal(err)}
	regionStats.Division, regionStats.DivisionRank = divisions.GetDivision(regionStats.LadderPoints)
}

func setClientDivision(t *testing.T, c *Client, div string){
	var set_div *Division;
	
	for d := range divisions {
		if(divisions[d].Name == div){
			set_div = divisions[d];
		}
	}

	points := int64((set_div.PromotionThreshold - set_div.DemotionThreshold)/2 + set_div.DemotionThreshold)
	c.LadderPoints = points

	updateClientDivisions(t,c)
}

func makeMM(t *testing.T) *Matchmaker{
	matchmaker = &Matchmaker{
		register:     make(chan *ClientConnection, 256),
		callback:     make(chan bool, 256),
		unregister:   make(chan *ClientConnection),
		participants: make(map[*ClientConnection]*MatchmakerParticipant),
		regionParticipants: map[BattleNetRegion]map[*ClientConnection]*MatchmakerParticipant{
			BATTLENET_REGION_NA:  make(map[*ClientConnection]*MatchmakerParticipant),
			BATTLENET_REGION_EU:  make(map[*ClientConnection]*MatchmakerParticipant),
			BATTLENET_REGION_KR:  make(map[*ClientConnection]*MatchmakerParticipant),
			BATTLENET_REGION_CN:  make(map[*ClientConnection]*MatchmakerParticipant),
			BATTLENET_REGION_SEA: make(map[*ClientConnection]*MatchmakerParticipant),
		},
		matchCache:             make(map[int64]*MatchmakerMatch),
		matchParticipantCache:  make(map[int64]*MatchmakerMatchParticipant),
		matchParticipantsCache: make(map[int64][]*MatchmakerMatchParticipant),
	}

	newmap := &Map {
		Id: 0,
		Region: region,
		SanitizedName: "testmap",
		InRankedPool: true,
	}

	// maps := make(Maps)
	maps[newmap.Id] = newmap 

	return matchmaker
}

func makeMatchDivs(t *testing.T, mm *Matchmaker, c1, c2 *Client, mp1, mp2 *MatchmakerParticipant, p1, p2 int64){
	c1.LadderPoints = p1
	updateClientDivisions(t,c1)

	c2.LadderPoints = p2
	updateClientDivisions(t,c2)

	mm.makeMatch(mp1,mp2,region)
	c1.Defeat(c2, region)
}

func TestSameMatch(t *testing.T) {
	mm := makeMM(t)
	c1 := createMockUser(t, 1)
	c2 := createMockUser(t, 2)

	mp1 := NewMatchmakerParticipant(c1);
	mp2 := NewMatchmakerParticipant(c2);

	mm.makeMatch(mp1,mp2,region)

	match1 := <-mp1.match
	match2 := <-mp2.match

	assert.Equal(t, match1.Id, match2.Id, "Player one and player two should be in the same match")
}

func TestStartDivisionPlacement(t *testing.T) {
	// mm := makeMM(t)
	c := createMockUser(t, 1)

	assert.Equal(t, ladderStartingPoints, c.client.LadderPoints, "Client should start with %v points", ladderStartingPoints)

	startingDivision, _ := divisions.GetDivision(ladderStartingPoints);
	assert.Equal(t, c.client.Division.Id, startingDivision.Id, "Client should start in division %v", startingDivision.Name)

	// Will usually be 1000 or 1200 points, we'll count on that (D rank)
}

func TestLadderGroups(t *testing.T){

}

func TestPointsWinD(t *testing.T) {
	mm := makeMM(t)
	c1 := createMockUser(t, 1)
	c2 := createMockUser(t, 2)

	mp1 := NewMatchmakerParticipant(c1);
	mp2 := NewMatchmakerParticipant(c2);

	// D win against D --> W: 100, L: -50
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 1200, 1200)
	assert.Equal(t, int64(1200 + 100), c1.client.LadderPoints, "D winning against D should win 100 points.")
	assert.Equal(t, int64(1200 - 50), c2.client.LadderPoints, "D losing against D should lose 50 points.")

	// D vs D- --> W: 75, L: -37
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 1200, 600)
	assert.Equal(t, int64(1200 + 75), c1.client.LadderPoints, "D winning against D- should win 100 points.")
	assert.Equal(t, int64(600 - 37), c2.client.LadderPoints, "D- losing against D should lose 50 points.")
}

func TestPointsWinA(t *testing.T){
	mm := makeMM(t)
	c1 := createMockUser(t, 1)
	c2 := createMockUser(t, 2)

	mp1 := NewMatchmakerParticipant(c1);
	mp2 := NewMatchmakerParticipant(c2);

	// A win against E --> W: 10, L: -0
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 11500, 250)
	assert.Equal(t, int64(11500 + 10), c1.client.LadderPoints, "A winning against E should win 10 points.")
	assert.Equal(t, int64(250 - 0), c2.client.LadderPoints, "E losing against A should lose 0 points.")

	// A win against D- --> W: 10, L: -0
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 11500, 550)
	assert.Equal(t, int64(11500 + 10), c1.client.LadderPoints, "A winning against D- should win 10 points.")
	assert.Equal(t, int64(550 - 0), c2.client.LadderPoints, "D- losing against A should lose 0 points.")

	// A win against D --> W: 10, L: -0
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 11500, 1200)
	assert.Equal(t, int64(11500 + 10), c1.client.LadderPoints, "A winning against D should win 10 points.")
	assert.Equal(t, int64(1200 - 0), c2.client.LadderPoints, "D losing against A should lose 0 points.")

	// A win against D+ --> W: 10, L: -0
	makeMatchDivs(t, mm, c1.client, c2.client, mp1, mp2, 11500, 2500)
	assert.Equal(t, int64(11500 + 10), c1.client.LadderPoints, "A winning against D+ should win 10 points.")
	assert.Equal(t, int64(2500 - 0), c2.client.LadderPoints, "D+ losing against A should lose 0 points.")
}