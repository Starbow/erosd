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

	MATCHMAKING_LONG_PROCESS_NOSHOW = 1
	MATCHMAKING_LONG_PROCESS_DRAW   = 2
)

var (
	matchmakingMatchTimeout int64 = 2 * 60 * 60
	// The time before long processes can be launched
	matchmakingLongProcessUnlockTime int64 = 60
	// The time that a long process takes.
	matchmakingLongProcessResponseTime int64 = 240
	// The minimum amount of points required for the user to be able to select search radius.
	matchmakingMinimumSearchRangePoints int64 = 750
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
	MapId    *int64
	AddTime  int64
	EndTime  int64
	Quality  float64
	Region   BattleNetRegion
	Channel  string
	ChatRoom string

	longProcessCount     int64     `db:"-"`
	longProcessType      int64     `db:"-"`
	longProcessActive    bool      `db:"-"`
	longProcessInitiator *Client   `db:"-"`
	longProcessResponse  chan bool `db:"-"`
}

type MatchmakerMatchParticipant struct {
	Id           int64
	MatchId      *int64
	ClientId     *int64
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

func (self *MatchmakerMatch) CanLongProcess() bool {
	return (time.Now().Unix() - self.AddTime) >= matchmakingLongProcessUnlockTime
}

func (self *MatchmakerMatch) EndMatch() {
	matchmaker.EndMatch(self.Id)
}

func (self *MatchmakerMatch) longProcessProc(initiator *Client, process int) {
	self.longProcessActive = false

	if process == MATCHMAKING_LONG_PROCESS_DRAW {
		matchmaker.logger.Println("Game", self.Id, "ended with no result")
		self.EndMatch()
	} else if process == MATCHMAKING_LONG_PROCESS_NOSHOW {
		matchmaker.logger.Println("Game", self.Id, "ended with walkover for", initiator.Id, initiator.Username)
		var opponent *Client
		if initiator.PendingMatchmakingOpponentId != nil {
			opponent = clientCache.Get(*initiator.PendingMatchmakingOpponentId)
		}
		if opponent == nil {
			self.EndMatch()
		} else {
			self.CreateForfeit(opponent)
		}
	}
}

func (self *MatchmakerMatch) StartLongProcess(initiator *Client, process int) bool {
	if !self.CanLongProcess() {
		return false
	}

	if self.longProcessActive {
		return false
	}

	matchmaker.logger.Println("Game", self.Id, "long process", process, "requested by", initiator.Id, initiator.Username)

	self.longProcessCount += 1
	self.longProcessActive = true
	self.longProcessInitiator = initiator
	self.longProcessResponse = make(chan bool)

	// If they're spamming the feature just abort the match.
	if self.longProcessCount == 3 {
		matchmaker.logger.Println("Game", self.Id, "long process spam aborted game")
		self.EndMatch()
		return true
	}

	go func() {
		timer := time.NewTimer(time.Second * time.Duration(matchmakingLongProcessResponseTime))

		select {
		case <-timer.C:
			self.longProcessProc(initiator, process)

		case response := <-self.longProcessResponse:
			matchmaker.logger.Println("Game", self.Id, "long process client response:", response)
			// Client has responded to us. Continue if they confirm the long process.
			if response {
				self.longProcessProc(initiator, process)
			}

			self.longProcessActive = false
			return
		}
	}()

	return true
}

func (mmm *MatchmakerMatch) CreateForfeit(client *Client) (result *MatchResult, players []*MatchResultPlayer, err error) {
	matchmaker.logger.Println("Forfeiting", client.Id, client.Username)
	participants := matchmaker.MatchParticipants(mmm.Id)
	result = &MatchResult{
		DateTime:          time.Now().Unix(),
		MapId:             mmm.MapId,
		MatchmakerMatchId: &mmm.Id,
		Region:            mmm.Region,
	}

	err = dbMap.Insert(result)
	if err != nil {
		matchmaker.logger.Println("Forfeit insert error", err)
		result = nil
		return
	}

	var opponentClient *Client
	var player, opponent MatchResultPlayer
	for x := range participants {
		if participants[x].ClientId != nil && *participants[x].ClientId != client.Id {
			opponentClient = clientCache.Get(*participants[x].ClientId)
		}
	}

	if opponentClient == nil {
		matchmaker.logger.Println("Opponent client not found")
		result = nil
		client.ForfeitMatchmadeMatch()
		err = ErrLadderPlayerNotFound

		return
	}

	player.MatchId = &result.Id
	player.ClientId = &client.Id
	player.Victory = false
	player.Race = "Forfeit"

	opponent.MatchId = &result.Id
	opponent.ClientId = &opponentClient.Id
	opponent.Victory = true
	opponent.Race = "Walkover"

	playerRegion, _ := client.RegionStats(mmm.Region)
	opponentRegion, _ := opponentClient.RegionStats(mmm.Region)

	if playerRegion == nil || opponentRegion == nil {
		player.PointsBefore = client.LadderPoints
		opponent.PointsBefore = opponentClient.LadderPoints
	} else {
		player.PointsBefore = playerRegion.LadderPoints
		opponent.PointsBefore = opponentRegion.LadderPoints
	}

	client.ForfeitMatchmadeMatch()

	if playerRegion == nil || opponentRegion == nil {
		player.PointsAfter = client.LadderPoints
		opponent.PointsAfter = opponentClient.LadderPoints
	} else {
		player.PointsAfter = playerRegion.LadderPoints
		opponent.PointsAfter = opponentRegion.LadderPoints
	}

	player.PointsDifference = player.PointsAfter - player.PointsBefore
	opponent.PointsDifference = opponent.PointsAfter - opponent.PointsBefore

	client.PendingMatchmakingId = nil
	client.PendingMatchmakingOpponentId = nil
	opponentClient.PendingMatchmakingId = nil
	opponentClient.PendingMatchmakingOpponentId = nil
	_, uerr := dbMap.Update(client, opponentClient)
	if uerr != nil {
		matchmaker.logger.Println(uerr)
		err = ErrDbInsert
		return
	}

	uerr = dbMap.Insert(&player, &opponent)
	if uerr != nil {
		matchmaker.logger.Println(uerr)
	}
	players = []*MatchResultPlayer{&player, &opponent}

	return
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
		match.longProcessActive = false
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

		res, err := dbMap.Select(&MatchmakerMatchParticipant{}, "SELECT * FROM matchmaker_match_participants WHERE MatchId=?", id)
		m = make([]*MatchmakerMatchParticipant, 0, len(res))
		if err != nil {
			return nil
		}

		for x := range res {
			m = append(m, res[x].(*MatchmakerMatchParticipant))
		}

		mm.matchParticipantsCache[id] = m
	}

	return m
}

func (mm *Matchmaker) EndMatch(id int64) {
	match := mm.Match(id)

	if match != nil {

		participants := mm.MatchParticipants(id)
		match.EndTime = time.Now().Unix()

		for x := range participants {
			if participants[x] == nil {
				log.Println("Nil participant for", id)
				continue
			}
			var client *Client = nil
			if participants[x].ClientId != nil {
				client = clientCache.Get(*participants[x].ClientId)
			}
			if client.PendingMatchmakingId != nil && *client.PendingMatchmakingId == id {
				client.PendingMatchmakingId = nil
				client.PendingMatchmakingOpponentId = nil
				client.PendingMatchmakingRegion = 0

				dbMap.Update(client)
				go client.Broadcast("MMI", nil)
			}

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
	vetoes1, _ := player1.connection.client.Vetoes()
	vetoes2, _ := player2.connection.client.Vetoes()
	selectedMap := maps.Random(player1.region, vetoes1, vetoes2)
	if selectedMap == nil {
		selectedMap = maps.Random(player1.region)
		if selectedMap == nil {
			log.Println("No map found while matching", player1.client.Username, player2.client.Username)
			go func() {
				player1.abort <- true
			}()
			go func() {
				player2.abort <- true
			}()
			return
		}
	}
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
	match.MapId = &selectedMap.Id
	match.Channel = battleNetChannel
	match.longProcessActive = false

	room := GetChatRoom(erosChatRoom, "", false, false)

	if room != nil {
		match.ChatRoom = erosChatRoom
	}

	err := dbMap.Insert(&match)
	mm.matchCache[match.Id] = &match

	if err == nil {
		var p1, p2 MatchmakerMatchParticipant
		p1time := time.Since(player1.enrollTime)
		p2time := time.Since(player2.enrollTime)
		p1.MatchId = &match.Id
		p2.MatchId = &match.Id
		p1.ClientId = &player1.connection.client.Id
		p2.ClientId = &player2.connection.client.Id
		p1.Points = player1.points
		p2.Points = player2.points
		p1.RatingMean = player1.connection.client.RatingMean
		p2.RatingMean = player2.connection.client.RatingMean
		p1.RatingStdDev = player1.connection.client.RatingStdDev
		p2.RatingStdDev = player2.connection.client.RatingStdDev
		p1.QueueTime = p1time.Seconds()
		p2.QueueTime = p2time.Seconds()
		err = dbMap.Insert(&p1, &p2)

		if err != nil {
			mm.matchParticipantCache[p1.Id] = &p1
			mm.matchParticipantCache[p2.Id] = &p2

			mm.matchParticipantsCache[match.Id] = []*MatchmakerMatchParticipant{&p1, &p2}
			matchmaker.logger.Println("Insert error", err)
		}
	} else {
		matchmaker.logger.Println("Insert failed", err)
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
