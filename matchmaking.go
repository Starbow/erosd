package main

import (
	"fmt"
	"github.com/ChrisHines/GoSkills/skills"
	"github.com/ChrisHines/GoSkills/skills/trueskill"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

// shut up the unused package warning
var _ = log.Ldate

//Matchmaking regions are for bitwise comparisons
const (
	MATCHMAKING_REGION_NA  = 1
	MATCHMAKING_REGION_EU  = 2
	MATCHMAKING_REGION_KR  = 4
	MATCHMAKING_REGION_SEA = 8
	MATCHMAKING_REGION_CN  = 16

	MATCHMAKING_TYPE_1V1 = 1
)

var (
	matchmakingMatchTimeout int64 = 2 * 60 * 60
)

type Matchmaker struct {
	// The actual matchmaker
	register               chan *ClientConnection
	callback               chan bool
	unregister             chan *ClientConnection
	participants           map[*ClientConnection]*MatchmakerParticipant
	regionParticipants     map[BattleNetRegion]map[*ClientConnection]*MatchmakerParticipant
	matchCache             map[int64]*MatchmakerMatch
	matchParticipantCache  map[int64]*MatchmakerMatchParticipant
	matchParticipantsCache map[int64][]*MatchmakerMatchParticipant
	logger                 *log.Logger
	logFile                *os.File
	sync.RWMutex
}

type MatchmakerParticipant struct {
	connection  *ClientConnection
	client      *Client
	enrollTime  time.Time // We track when this started, so
	team        skills.Team
	points      int64
	radius      int64 // x * points per division
	region      BattleNetRegion
	queueType   int64 // 1v1, 2v2
	match       chan *MatchmakerMatch
	abort       chan bool
	matching    bool
	vetoes      []*Map
	opponent    *MatchmakerParticipant
	selectedMap *Map
}

type MatchmakerPotentialMatch struct {
	opponent        *MatchmakerParticipant
	pointDifference int64
}

type MatchmakerMatch struct {
	Id       int64
	MapId    int64
	AddTime  int64
	EndTime  int64
	Quality  float64
	Region   BattleNetRegion
	Channel  string
	ChatRoom string
}

type MatchmakerMatchParticipant struct {
	Id           int64
	MatchId      int64
	ClientId     int64
	Points       int64
	RatingMean   float64
	RatingStdDev float64
	QueueTime    float64
}

func initMatchmaking() {
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

	var logfile string = path.Join(logPath, fmt.Sprintf("%d-mm.log", os.Getpid()))
	file, err := os.Create(logfile)
	if err != nil {
		log.Println("Failed to create log file", logfile, "for matchmaker")
		matchmaker.logger = log.New(os.Stdout, fmt.Sprintf("mm"), log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		log.Println("Logging matchmaker to", logfile)
		matchmaker.logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
		matchmaker.logFile = file
	}

	go matchmaker.run()
}

func NewMatchmakerParticipant(connection *ClientConnection) *MatchmakerParticipant {
	region, _ := connection.client.RegionStats(connection.client.LadderSearchRegion)

	// TrueSkill stuff
	team := skills.NewTeam()
	player := skills.NewPlayer(connection.client.Id)
	team.AddPlayer(*player, skills.NewRating(connection.client.RatingMean, connection.client.RatingStdDev))
	var points int64
	if region == nil {
		team.AddPlayer(*player, skills.NewRating(connection.client.RatingMean, connection.client.RatingStdDev))
		points = connection.client.LadderPoints
	} else {
		team.AddPlayer(*player, skills.NewRating(region.RatingMean, region.RatingStdDev))
		points = region.LadderPoints
	}

	return &MatchmakerParticipant{
		connection: connection,
		client:     connection.client,
		enrollTime: time.Now(),
		team:       team,
		points:     points,
		radius:     connection.client.LadderSearchRadius,
		region:     connection.client.LadderSearchRegion,
		matching:   false,
		abort:      make(chan bool),
		match:      make(chan *MatchmakerMatch),
	}
}
func (mm *Matchmaker) Match(id int64) *MatchmakerMatch {

	match, ok := mm.matchCache[id]

	if !ok {
		match = &MatchmakerMatch{}
		err := dbMap.SelectOne(match, "SELECT * FROM matchmaker_matches WHERE Id=? LIMIT 1", id)

		if err != nil || match.Id == 0 {
			return nil
		}

		mm.matchCache[id] = match
	}

	return match
}

func (mm *Matchmaker) MatchParticipant(id int64) *MatchmakerMatchParticipant {

	m, ok := mm.matchParticipantCache[id]

	if !ok {
		m = &MatchmakerMatchParticipant{}
		err := dbMap.SelectOne(m, "SELECT * FROM matchmaker_match_participants WHERE Id=? LIMIT 1", id)

		if err != nil || m.Id == 0 {
			return nil
		}

		mm.matchParticipantCache[id] = m
	}

	return m
}

func (mm *Matchmaker) MatchParticipants(id int64) []*MatchmakerMatchParticipant {

	m, ok := mm.matchParticipantsCache[id]

	if !ok {

		res, err := dbMap.Select(&MatchmakerMatchParticipant{}, "SELECT * FROM matchmaker_match_participants WHERE MatchId=? LIMIT 1", id)
		m = make([]*MatchmakerMatchParticipant, len(res))
		if err != nil {
			return nil
		}

		for x := range res {
			m[x] = res[x].(*MatchmakerMatchParticipant)
		}

		mm.matchParticipantsCache[id] = m
	}

	return m
}

func (mm *Matchmaker) EndMatch(id int64, participant ...*Client) {
	match := mm.Match(id)

	if match != nil {
		match.EndTime = time.Now().Unix()

		for x := range participant {
			if participant[x] == nil {
				continue
			}

			participant[x].PendingMatchmakingId = 0
			participant[x].PendingMatchmakingOpponentId = 0
			participant[x].PendingMatchmakingRegion = 0

		}

		dbMap.Update(match)

	}

}

//Match 2 players against each other.
func (mm *Matchmaker) makeMatch(player1 *MatchmakerParticipant, player2 *MatchmakerParticipant) {
	quality := player1.Quality(player2)
	go func() {
		mm.unregister <- player1.connection
		mm.unregister <- player2.connection
	}()

	selectedMap := maps.Random(player1.region, player1.connection.client.Vetoes(), player2.connection.client.Vetoes())
	battleNetChannel := fmt.Sprintf("eros%d%d%d%d", player1.region, player1.client.Id, player2.client.Id, rand.Intn(99))
	erosChatRoom := cleanChatRoomName(fmt.Sprintf("MM%d%d%d", player1.region, player1.client.Id, player2.client.Id))

	player1.opponent = player2
	player2.opponent = player1
	player1.selectedMap = selectedMap
	player2.selectedMap = selectedMap

	var match MatchmakerMatch
	match.AddTime = time.Now().Unix()
	match.Quality = quality
	match.Region = player1.region
	match.MapId = selectedMap.Id
	match.Channel = battleNetChannel

	room, ok := chatRooms[erosChatRoom]
	var err error
	if !ok {
		room, err = NewChatRoom(erosChatRoom, "", false, false)
		if err != nil {
			log.Println("Error creating matchmaking chat", err, erosChatRoom)
		}
	}

	if room != nil {
		match.ChatRoom = erosChatRoom
	}

	err = dbMap.Insert(&match)
	mm.matchCache[match.Id] = &match

	if err == nil {
		var p1, p2 MatchmakerMatchParticipant
		p1.MatchId = match.Id
		p2.MatchId = match.Id
		p1.ClientId = player1.connection.client.Id
		p2.ClientId = player1.connection.client.Id
		p1.Points = player1.points
		p2.Points = player2.points
		p1.RatingMean = player1.connection.client.RatingMean
		p2.RatingMean = player1.connection.client.RatingMean
		p1.RatingStdDev = player1.connection.client.RatingStdDev
		p2.RatingStdDev = player2.connection.client.RatingStdDev
		p1.QueueTime = time.Since(player1.enrollTime).Seconds()
		p2.QueueTime = time.Since(player2.enrollTime).Seconds()
		err = dbMap.Insert(&p1, &p2)

		if err != nil {
			mm.matchParticipantCache[p1.Id] = &p1
			mm.matchParticipantCache[p2.Id] = &p2

			mm.matchParticipantsCache[match.Id] = []*MatchmakerMatchParticipant{&p1, &p2}
		}
	}

	go func() {
		player1.match <- &match
	}()
	go func() {
		player2.match <- &match
	}()

}

// Matchmaking worker
func (mm *Matchmaker) run() {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for {
			//Go primer:
			//Select works similarly to switch syntactically, but uses channels.
			//Whichever route presents a value first will be executed.
			//A channel is (in this case) a fifo that links goroutines.
			select {
			case <-ticker.C:
				// Maintain a list of participants we've already looped through.
				// Any comparisons against them will be duplicate.

				//TODO: We can use the regional map if required in the future.
				complete := make([]*MatchmakerParticipant, 0, len(mm.participants))
				for k, v := range mm.participants {
					if v.matching {
						continue
					}

					potentials := make([]MatchmakerPotentialMatch, 0, len(mm.participants))

				outer:
					for l, w := range mm.participants {

						// Check we're not comparing ourself
						if k == l {
							continue
						}

						// Check we're not the same client.
						if k.client.Id == l.client.Id {
							continue
						}

						// Check we're not already matched.
						if w.matching {
							continue
						}

						//Check against overall complete participants and skip them
						for _, x := range complete {
							if x == w {
								// Continue the loop above this one
								continue outer
							}
						}

						//log.Println("Compare", k.client.Id, "to", l.client.Id, "quality", quality, b1, b2)

						if v.IsMatch(w) {
							potentials = append(potentials, MatchmakerPotentialMatch{opponent: w,
								pointDifference: int64(math.Abs(float64(v.points - w.points))),
							})
						}
					}

					// If we have potential matches, find the lowest difference and match them.

					if len(potentials) > 0 {
						sort.Sort(ByRatingDifference(potentials))
						x := potentials[0].opponent

						v.matching = true
						x.matching = true
						go mm.makeMatch(v, x)
					}

					//Mark our scent
					complete = append(complete, v)
				}
			case client := <-mm.register:
				delete(mm.participants, client)
				delete(mm.regionParticipants[BATTLENET_REGION_NA], client)
				delete(mm.regionParticipants[BATTLENET_REGION_EU], client)
				delete(mm.regionParticipants[BATTLENET_REGION_KR], client)
				delete(mm.regionParticipants[BATTLENET_REGION_CN], client)
				delete(mm.regionParticipants[BATTLENET_REGION_SEA], client)

				mm.participants[client] = NewMatchmakerParticipant(client)
				mm.regionParticipants[client.client.LadderSearchRegion][client] = mm.participants[client]
				go func() {
					mm.callback <- true
				}()
			case client := <-mm.unregister:
				delete(mm.participants, client)
				delete(mm.regionParticipants[BATTLENET_REGION_NA], client)
				delete(mm.regionParticipants[BATTLENET_REGION_EU], client)
				delete(mm.regionParticipants[BATTLENET_REGION_KR], client)
				delete(mm.regionParticipants[BATTLENET_REGION_CN], client)
				delete(mm.regionParticipants[BATTLENET_REGION_SEA], client)
			}

		}
	}()
}

func (mp *MatchmakerParticipant) Quality(opponent *MatchmakerParticipant) float64 {
	teams := []skills.Team{mp.team, opponent.team}
	var calc trueskill.TwoPlayerCalc
	quality := calc.CalcMatchQual(skills.DefaultGameInfo, teams)

	return quality
}

// Worst math ever
func (mp *MatchmakerParticipant) SearchBoundaries() (lower, upper, variance int64) {
	var (
		elapsed              = float64(time.Since(mp.enrollTime).Seconds())
		participants float64 = float64(len(matchmaker.regionParticipants[mp.region]))
		r            int64
	)

	if participants < 20 {
		r = int64(12 + (elapsed * 12))
	} else if participants < 140 {
		r = int64(15 + (200*elapsed)/participants)
	} else {
		r = int64(15 + (2 * elapsed))
	}

	if mp.radius > 0 {
		cap := mp.radius * divisionPoints
		if r > cap {
			r = cap
		}
	}
	return mp.points - r, mp.points + r, r
}

func (mp *MatchmakerParticipant) IsMatch(mp2 *MatchmakerParticipant) bool {
	if mp.region != mp2.region {
		return false
	}

	r1 := mp.points
	r2 := mp2.points
	r1l, r1u, r1v := mp.SearchBoundaries()
	r2l, r2u, r2v := mp2.SearchBoundaries()

	r1Match := (r1+r1v >= r2l && r1-r1v <= r2u)
	r2Match := (r2+r2v >= r1l && r2-r2v <= r1u)
	return (r1Match && r2Match)
}

type ByRatingDifference []MatchmakerPotentialMatch

func (a ByRatingDifference) Len() int           { return len(a) }
func (a ByRatingDifference) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRatingDifference) Less(i, j int) bool { return a[i].pointDifference < a[j].pointDifference }
