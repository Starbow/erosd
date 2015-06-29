package main

import (
	"errors"
	"github.com/ChrisHines/GoSkills/skills"
	"github.com/ChrisHines/GoSkills/skills/trueskill"
	"github.com/Starbow/erosd/buffers"
	"log"
	"math/rand"
	"strings"
)

var _ = log.Ldate
var (
	ErrLadderPlayerNotFound           = errors.New("The player was not found in the database.")
	ErrLadderClientNotInvolved        = errors.New("None of the client's registered characters were found in the replay participant list.")
	ErrLadderInvalidMatchParticipents = errors.New("All participents of a game must be registered.")
	ErrLadderInvalidMap               = errors.New("Matches must be on a valid map in the map pool.")
	ErrLadderInvalidFormat            = errors.New("Matches must be a 1v1 with no observers.")
	ErrLadderDuplicateReplay          = errors.New("The provided has been processed previously.")
	ErrLadderGameTooShort             = errors.New("The provided game was too short.")
	ErrLadderWrongOpponent            = errors.New("The provided game was not against your matchmade opponent. You have been forfeited.")
	ErrLadderWrongMap                 = errors.New("The provided game was not on the correct map.")
	ErrLadderWrongSpeed               = errors.New("The provided game was not on the Faster speed setting.")
	ErrLadderGameNotPrearranged       = errors.New("The provided game was not arranged by the Eros matchmaker.")

	ladder = Iccup{}
)

type Division struct {
	Id                 int64   `db:"id"`
	Name               string  `db:"name"`
	PromotionThreshold float64 `db:"promotion_threshold"`
	DemotionThreshold  float64 `db:"demotion_threshold"`
	IconUrl            string  `db:"icon_url"`
	SmallIconUrl       string  `db:"small_icon_url"`
	LadderGroup        int64   `db:"ladder_group"`
	System             string  `db:"system"`
}

func (d *Division) DivisionMessage() *protobufs.Division {
	var division protobufs.Division
	division.Name = &d.Name
	division.PromotionThreshold = &d.PromotionThreshold
	division.DemotionThreshold = &d.DemotionThreshold
	division.Id = &d.Id
	division.IconUrl = &d.IconUrl
	division.SmallIconUrl = &d.SmallIconUrl

	return &division
}

type Divisions []*Division
type Maps map[int64]*Map

var (
	divisionNames []string = []string{"E", "D", "C", "B", "A"}
	divisions     Divisions
	maps          Maps = Maps{}

	divisionCount             int64
	divisionIncrements        float64
	divisionFirstRating       float64
	ladderStartingPoints      int64             = 0
	ladderWinPointsBase       int64             = 100
	ladderLosePointsBase      int64             = 50
	ladderWinPointsIncrement  float64           = 25
	ladderLosePointsIncrement float64           = 12.5
	ladderMaxMapVetos         int64             = 3
	ladderActiveRegions       []BattleNetRegion = []BattleNetRegion{BATTLENET_REGION_NA, BATTLENET_REGION_EU, BATTLENET_REGION_KR}
)

// Load maps from the database
func loadMaps() {
	results, err := dbMap.Select(&Map{}, "SELECT * FROM maps")
	newMaps := make(Maps)

	if err == nil {
		for x := range results {
			mapObject := results[x].(*Map)
			newMaps[mapObject.Id] = mapObject
			mapObject.SanitizedName = strings.TrimSpace(strings.ToLower(mapObject.BattleNetName))
		}
	} else {
		log.Panic("Error loading maps", err)
	}

	maps = newMaps
}

// Create divisions. There will be [subdivisionCount] subdivisions per division.
// The final division will only have one subdivision.
func initDivisions() {
	divisions = make(Divisions, 0, divisionCount)

	div, err := dbMap.Select(Division{}, "SELECT * FROM divisions ORDER BY promotion_threshold")
	if err != nil {
		panic(err)
	}

	if len(div) > 0 {
		for x := range div {
			divisions = append(divisions, div[x].(*Division))
		}
	} else {
		i := int64(0)
		for {
			if i > divisionCount {
				break
			}
			var rating float64
			if i == 0 {
				rating = 0
			} else {
				rating = divisionFirstRating + (float64(i-1) * divisionIncrements)
			}

			divisions = append(divisions, &Division{
				PromotionThreshold: rating,
				DemotionThreshold:  rating - 1,
				Name:               divisionNames[i],
				Id:                 0,
				LadderGroup:        i,
			})

			i++
		}

		for _, x := range divisions {
			err = dbMap.Insert(x)
			if err != nil {
				panic(err)
			}
		}
	}

	return
}

func (this *Division) GetDifference(other *Division) int64 {
	if this == nil || other == nil {
		return 0
	}
	var (
		p1, p2 int64
		i      = int64(len(divisions))
	)

	for {
		i--

		if divisions[i] == this {
			p1 = i
		} else if divisions[i] == other {
			p2 = i
		}

		if i == 0 {
			break
		}
	}

	return p2 - p1
}

// TODO: add to Ladderer interface
func (d Divisions) GetDivision(points int64) (division *Division, position int64) {
	i := int64(len(d))
	for {
		i--

		if float64(points) >= float64(d[i].PromotionThreshold) {
			division = d[i]
			position = i
			break
		}

		if i == 0 {
			break
		}
	}

	log.Println("GetDivision:", division)

	return nil, 0
}

// Get the difference in ranks
// DEPRECATED: using interface now
// func (d Divisions) GetDifference(mmr, mmr2 float64) int64 {
// 	_, p1 := d.GetDivision(mmr)
// 	_, p2 := d.GetDivision(mmr2)

// 	return p2 - p1
// }

func (m Maps) Get(region BattleNetRegion, name string) *Map {
	sanitized := strings.TrimSpace(strings.ToLower(name))
	for x := range m {
		if m[x].Region != region {
			continue
		}

		if m[x].SanitizedName == sanitized {
			return m[x]
		}
	}

	return nil
}

func (m Maps) GetId(region BattleNetRegion, id int) *Map {
	for x := range m {
		if m[x].Region != region {
			continue
		}

		if m[x].BattleNetID == id {
			return m[x]
		}
	}

	return nil
}

func (m Maps) Random(region BattleNetRegion, veto ...[]*Map) *Map {
	var pool []*Map = make([]*Map, 0, 5)

mapLoop:
	for x := range m {
		if m[x].Region != region || !m[x].InRankedPool {
			continue
		}

		// Check vetoes, and continue main loop if found
		for y := range veto {
			for z := range veto[y] {
				if veto[y][z] == nil {
					continue
				}

				if m[x].BattleNetID == veto[y][z].BattleNetID && m[x].Region == veto[y][z].Region {
					continue mapLoop
				}
			}
		}

		pool = append(pool, m[x])
	}

	if len(pool) == 0 {
		return nil
	}

	return pool[rand.Intn(len(pool))]

}

func (self Maps) MapPoolMessage() *protobufs.MapPool {
	var pool []*Map = make([]*Map, 0, 12*len(ladderActiveRegions))
	for x := range self {
		if self[x].InRankedPool {
			pool = append(pool, self[x])
		}
	}

	mapPoolMessage := &protobufs.MapPool{
		Map: make([]*protobufs.Map, 0, len(pool)),
	}

	for x := range pool {
		mapPoolMessage.Map = append(mapPoolMessage.Map, pool[x].MapMessage())
	}

	return mapPoolMessage
}

type Map struct {
	Id            int64
	Region        BattleNetRegion
	BattleNetID   int
	BattleNetName string
	InRankedPool  bool
	Description   string
	InfoUrl       string
	PreviewUrl    string
	SanitizedName string `db:"-"`
}

func (m *Map) MapMessage() *protobufs.Map {
	var (
		msg    protobufs.Map
		region protobufs.Region = protobufs.Region(m.Region)
		id     int32            = int32(m.BattleNetID)
	)
	msg.Region = &region
	msg.BattleNetName = &m.BattleNetName
	msg.BattleNetId = &id
	msg.Description = &m.Description
	msg.InfoUrl = &m.InfoUrl
	msg.PreviewUrl = &m.PreviewUrl
	return &msg
}

type MapVeto struct {
	Id       int64
	ClientId int64
	MapId    int64
}

type MatchResult struct {
	Id                int64
	MapId             *int64 // Map
	MatchmakerMatchId *int64
	DateTime          int64 // unix
	Region            BattleNetRegion
}

type MatchResultPlayer struct {
	Id               int64
	MatchId          *int64
	ClientId         *int64
	CharacterId      *int64
	PointsBefore     int64
	PointsAfter      int64
	PointsDifference int64
	Race             string
	Victory          bool
}

type MatchResultSource struct {
	Id         int64
	MatchId    *int64
	ReplayHash string
}

func NewMatchResult(replay *Replay, client *Client) (result *MatchResult, players []*MatchResultPlayer, err error) {
	// Find the local character
	// Find the opponent
	matchmaker.logger.Println("New replay from", client.Id, client.Username, *replay)
	var region BattleNetRegion
	if replay.Region != "" {
		region = ParseBattleNetRegion(replay.Region)
	} else if replay.Gateway != "" {
		region = ParseBattleNetRegion(replay.Gateway)
	}
	if replay.Speed != "Faster" {
		err = ErrLadderWrongSpeed
		matchmaker.logger.Println("Wrong speed from", client.Username)
		return
	}

	if replay.GameLength < 120 {
		err = ErrLadderGameTooShort
		matchmaker.logger.Println("Too short from", client.Username)
		return
	}

	m := maps.Get(region, replay.MapName)

	if m == nil {
		matchmaker.logger.Println("Map not found in pool", region, replay.MapName)
		err = ErrLadderInvalidMap
		return
	}
	if !m.InRankedPool {
		matchmaker.logger.Println("Map not found in ranked pool", region, replay.MapName)
		err = ErrLadderInvalidMap
		return
	}

	if len(replay.Observers) > 0 || len(replay.Players) != 2 {
		matchmaker.logger.Println("Invalid game format", len(replay.Observers), "observers", len(replay.Players), "players")
		err = ErrLadderInvalidFormat
		return
	}

	count, err := dbMap.SelectInt("SELECT COUNT(*) FROM match_result_sources WHERE ReplayHash=?", replay.Filehash)
	if err == nil && count > 0 {
		matchmaker.logger.Println("Duplicate replay")
		err = ErrLadderDuplicateReplay
		return
	}

	var (
		player   *MatchResultPlayer
		opponent *MatchResultPlayer
	)
	for x := range replay.Players {
		mrp, merr := NewMatchResultPlayer(replay, &replay.Players[x])
		if merr != nil || mrp == nil {
			matchmaker.logger.Println("MRP Error or nil", merr)
			err = ErrLadderInvalidMatchParticipents
			return
		}

		if mrp.ClientId != nil && *mrp.ClientId == client.Id {
			player = mrp
		} else {
			opponent = mrp
		}
	}

	//We don't have the player that submitted the replay.
	if player == nil {
		matchmaker.logger.Println("ErrLadderClientNotInvolved")
		err = ErrLadderClientNotInvolved
		return
	}

	//We don't have an opponent.
	if opponent == nil {
		matchmaker.logger.Println("ErrLadderInvalidMatchParticipents")
		err = ErrLadderInvalidMatchParticipents
		return
	}

	//Make sure we only have one and only one victor.
	if player.Victory == opponent.Victory {
		matchmaker.logger.Println("player.Victory == opponent.Victory")
		err = ErrLadderInvalidFormat
		return
	}

	if client.PendingMatchmakingId == nil || *client.PendingMatchmakingId == 0 {
		matchmaker.logger.Println("Invalid match.")
		err = ErrLadderGameNotPrearranged
		return
	}
	// We're only going to accept replays from the victor.
	// In future this should be changed to match games against the start time
	// I made a nice client-ID based mutex manager just for this purpose.
	//clientLockouts.LockIds(player.ClientId, opponent.ClientId)
	//clientLockouts.UnlockIds(player.ClientId, opponent.ClientId)
	if player.Victory {
		var opponentClient *Client = nil
		if opponent.ClientId != nil {
			opponentClient = clientCache.Get(*opponent.ClientId)
		}

		if opponentClient == nil {
			err = ErrLadderPlayerNotFound // new error for lookup fail
			matchmaker.logger.Println("Opponenet client nil", opponent.ClientId)
			return
		}

		if !client.IsMatchedWith(opponentClient) {
			client.ForfeitMatchmadeMatch()
			err = ErrLadderWrongOpponent
			matchmaker.logger.Println("Not matched with opponent", client.Id, client.Username)
			return
		}

		if !opponentClient.IsMatchedWith(client) {
			opponentClient.ForfeitMatchmadeMatch()
			matchmaker.logger.Println("Not matched with opponent", opponentClient.Id, opponentClient.Username)
		}

		var res MatchResult
		res.DateTime = replay.UnixTimestamp
		res.MapId = &m.Id

		if client.PendingMatchmakingId != nil {
			pendingMMID := *client.PendingMatchmakingId

			res.MatchmakerMatchId = &pendingMMID
		} else {
			res.MatchmakerMatchId = nil
		}

		res.Region = region

		err = dbMap.Insert(&res)
		if err != nil {
			matchmaker.logger.Println(err)
			err = ErrDbInsert
			return
		}

		var source MatchResultSource
		source.MatchId = &res.Id
		source.ReplayHash = replay.Filehash
		dbMap.Insert(&source)

		if !client.IsOnMap(m.Id) {
			err = ErrLadderWrongMap
			matchmaker.logger.Println("Wrong map, expected", m.Id)
			return
		}

		player.MatchId = &res.Id
		opponent.MatchId = &res.Id

		playerRegion, _ := client.RegionStats(region)
		opponentRegion, _ := opponentClient.RegionStats(region)

		// Fetch region stats for stats purposes.
		// Use global if we can't fetch them.

		if playerRegion == nil || opponentRegion == nil {
			player.PointsBefore = client.LadderPoints
			opponent.PointsBefore = opponentClient.LadderPoints
		} else {
			player.PointsBefore = playerRegion.LadderPoints
			opponent.PointsBefore = opponentRegion.LadderPoints
		}

		// Adjust points
		client.Defeat(opponentClient, region)

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
			err = ErrDbInsert
			matchmaker.logger.Println(uerr)
			return
		}

		uerr = dbMap.Insert(player, opponent)
		if uerr != nil {
			matchmaker.logger.Println(uerr)
		}
		result = &res
		players = []*MatchResultPlayer{player, opponent}

		go func() {
			client.BroadcastStatsMessage()
			client.BroadcastMatchmakingIdle()
		}()
		go func() {
			opponentClient.BroadcastStatsMessage()
			opponentClient.BroadcastMatchmakingIdle()
		}()
	}

	return
}

// DEPRECATED: Old system, should move to interface
func calculateNewPoints(winner, loser int64, winnerDivision, loserDivision *Division) (winnerNew, loserNew int64) {
	// Update points
	// GetDifference(2000, 1000) would return -1
	// GetDifference(2000, 3000) would return 1

	difference := winnerDivision.GetDifference(loserDivision)
	increase := ladderWinPointsBase + int64((ladderWinPointsIncrement * float64(difference)))
	decrease := ladderLosePointsBase - (int64((ladderLosePointsIncrement * float64(difference))) * -1)

	if increase < 1 {
		increase = 10
	}

	if decrease < 0 {
		decrease = 0
	}

	winnerNew = winner + increase
	loserNew = loser - decrease

	if winnerNew < 0 {
		winnerNew = 0
	}

	if loserNew < 0 {
		loserNew = 0
	}

	return
}

func calculateNewRating(winnerId, loserId int64, winnerRating, winnerStdDev, loserRating, loserStdDev float64) (winnerNewRating, winnerNewStdDev, loserNewRating, loserNewStdDev, quality float64) {
	team1 := skills.NewTeam()
	team2 := skills.NewTeam()

	team1.AddPlayer(winnerId, skills.NewRating(winnerRating, winnerStdDev))
	team2.AddPlayer(loserId, skills.NewRating(loserRating, loserStdDev))

	teams := []skills.Team{team1, team2}

	var calc trueskill.TwoPlayerCalc
	ratings := calc.CalcNewRatings(skills.DefaultGameInfo, teams, 1, 2)
	quality = calc.CalcMatchQual(skills.DefaultGameInfo, teams)

	return ratings[winnerId].Mean(), ratings[winnerId].Stddev(), ratings[loserId].Mean(), ratings[loserId].Stddev(), quality
}

func NewMatchResultPlayer(replay *Replay, player *ReplayPlayer) (matchplayer *MatchResultPlayer, err error) {
	region, subregion, id, _ := ParseBattleNetProfileUrl(player.Url)
	character := characterCache.Get(region, subregion, id)

	if character == nil {
		return nil, ErrLadderPlayerNotFound
	}

	var mrp MatchResultPlayer

	mrp.CharacterId = &character.Id
	mrp.ClientId = character.ClientId
	mrp.Race = player.GameRace
	mrp.Victory = player.Victory == "Win"

	return &mrp, nil
}

func (mr *MatchResult) MatchResultMessage(players []*MatchResultPlayer) *protobufs.MatchResult {
	var (
		message    protobufs.MatchResult
		region     protobufs.Region = protobufs.Region(mr.Region)
		mapMessage *protobufs.Map   = maps[*mr.MapId].MapMessage()
	)

	message.Region = &region
	message.Map = mapMessage
	message.Participant = make([]*protobufs.MatchParticipant, 0, len(players))

	for x := range players {
		message.Participant = append(message.Participant, players[x].MatchParticipantMessage())
	}

	return &message
}

func (mrp *MatchResultPlayer) MatchParticipantMessage() *protobufs.MatchParticipant {
	var (
		message   protobufs.MatchParticipant
		client    *Client             = nil
		character *BattleNetCharacter = nil
	)
	if mrp.ClientId != nil {
		client = clientCache.Get(*mrp.ClientId)
	}
	if mrp.CharacterId != nil {
		character = characterCache.GetId(*mrp.CharacterId)
	}

	message.PointsBefore = &mrp.PointsBefore
	message.PointsAfter = &mrp.PointsAfter
	message.PointsDifference = &mrp.PointsDifference
	message.Victory = &mrp.Victory
	message.Race = &mrp.Race

	if client != nil {
		message.User = client.UserStatsMessage()
	}

	if character != nil {
		message.Character = character.CharacterMessage()
	}

	return &message
}
