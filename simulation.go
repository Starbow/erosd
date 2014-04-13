package main

import (
	"fmt"
	"github.com/Starbow/erosd/buffers"
	"log"
	"math/rand"
	"time"
)

type SimulatedUser struct {
	client     *Client
	conn       *ClientConnection
	characters []*BattleNetCharacter
}

func NewSimulatedUser(number, points, radius int64) *SimulatedUser {
	region_choices := []BattleNetRegion{
		BattleNetRegion(1),
		BattleNetRegion(2),
		//BattleNetRegion(3),
		//BattleNetRegion(5),
		//BattleNetRegion(6),
	}

	// First get the corresponding website user
	username := fmt.Sprintf("SimulatedUser%d", number)
	authtoken := fmt.Sprintf("AuthToken%d", number)
	site_user := GetRealUser(username, authtoken)
	log.Println("Spawning user", number, "using site_user", site_user.Id)

	// Then get the client, create if missing
	client := clientCache.Get(site_user.Id)
	if client == nil {
		log.Println("Creating new client for user", username)
		client = NewClient(site_user.Id)
		client.Username = site_user.Username
		client.LadderPoints = points
		client.LadderSearchRadius = radius
		err := dbMap.Insert(client)
		if err != nil {
			log.Println(err)
		}
	}

	// Create a character for this client with a random region
	region := region_choices[rand.Intn(len(region_choices))]
	character := characterCache.Get(region, 1, int(site_user.Id))
	if character == nil {
		log.Println("Creating new character for user", username, "in region", region)
		character = NewBattleNetCharacter(region, 1, int(site_user.Id), site_user.Username)
		character.IsVerified = true
		character.ClientId = &client.Id
		err := dbMap.Insert(character)
		if err != nil {
			log.Println(err)
		}
	}

	// Create a dummy connection for the client
	connection := &ClientConnection{
		id:            site_user.Id,
		conn:          nil,
		client:        client,
		authenticated: true,
	}

	// Return a fully packed SimulatedUser
	return &SimulatedUser{
		client:     client,
		conn:       connection,
		characters: []*BattleNetCharacter{character},
	}
}

func get_random_region_character(client *Client, region BattleNetRegion) *BattleNetCharacter {
	res, err := dbMap.Select(&BattleNetCharacter{}, "SELECT * FROM battle_net_characters WHERE ClientId=? AND Region=? LIMIT 1", client.Id, region)
	if err != nil {
		log.Println(err)
		return nil
	}
	return res[rand.Intn(len(res))].(*BattleNetCharacter)
}

func get_random_character(client *Client) *BattleNetCharacter {
	res, err := dbMap.Select(&BattleNetCharacter{}, "SELECT * FROM battle_net_characters WHERE ClientId=? LIMIT 1", client.Id)
	if err != nil {
		log.Println(err)
		return nil
	}
	return res[rand.Intn(len(res))].(*BattleNetCharacter)
}

func (user *SimulatedUser) Run() {
	races := []string{"Terran", "Zerg", "Protoss"}
	for {
		// Pick a random character to queue with
		character := get_random_character(user.client)
		user.client.LadderSearchRegions = []BattleNetRegion{character.Region}

		// Join the matchmaking queue
		matchmaker.register <- user.conn

		// Wait until we have been queued
		<-matchmaker.callback

		el, ok := matchmaker.participants[user.conn]
		if ok {
			// Wait until a match is found
			match := <-el.match
			opponent := el.opponent

			// Update the database record for the client
			user.client.PendingMatchmakingId = &match.Id
			user.client.PendingMatchmakingOpponentId = &opponent.client.Id
			user.client.PendingMatchmakingRegion = int64(match.Region)
			dbMap.Update(user.client)

			// Create new match making result record
			elapsed := int64(time.Since(el.enrollTime).Seconds())
			var mm_result protobufs.MatchmakingResult
			mm_result.Channel = &match.Channel
			mm_result.Quality = &match.Quality
			mm_result.ChatRoom = &match.ChatRoom
			mm_result.Timespan = &elapsed
			mm_result.Opponent = opponent.client.UserStatsMessage()
			mm_result.Map = el.selectedMap.MapMessage()

			// Force the highest id client to decide the match result
			if user.client.Id > opponent.connection.client.Id {
				log.Printf("%s (%d) found %s (%d) after %d seconds", user.client.Username, user.client.LadderPoints, opponent.connection.client.Username, opponent.connection.client.LadderPoints, elapsed)
				result := rand.Intn(100)
				if result < 50 {
					match.CreateForfeit(user.client)
				} else if result > 50 {
					match.CreateForfeit(opponent.client)
				} else {
					// Insert a match result record for this match
					bnet_map := maps.Get(character.Region, *mm_result.Map.BattleNetName)
					match_result := &MatchResult{
						DateTime:          time.Now().Unix(),
						MapId:             &bnet_map.Id,
						MatchmakerMatchId: &match.Id,
						Region:            user.client.LadderSearchRegions[0],
					}
					err := dbMap.Insert(match_result)
					if err != nil {
						log.Println(err)
					}

					// Get region stats for us and them
					userRegion, _ := user.client.RegionStats(character.Region)
					opponentRegion, _ := opponent.client.RegionStats(character.Region)

					// Create player record for us
					userPlayer := &MatchResultPlayer{
						MatchId:     &match_result.Id,
						ClientId:    &user.client.Id,
						CharacterId: &character.Id,
						Race:        races[rand.Intn(len(races))],
					}

					// Create player record for them
					opponentPlayer := &MatchResultPlayer{
						MatchId:     &match_result.Id,
						ClientId:    &opponent.client.Id,
						CharacterId: &get_random_region_character(opponent.client, match_result.Region).Id,
						Race:        races[rand.Intn(len(races))],
					}

					userPlayer.PointsBefore = userRegion.LadderPoints
					opponentPlayer.PointsBefore = opponentRegion.LadderPoints

					// Assign results for each player
					result := rand.Intn(100)
					if result < 50 {
						// User win
						userPlayer.Victory = true
						opponentPlayer.Victory = false

						if result < 3 {
							// User walkover
							opponent.client.ForfeitMatchmadeMatch()
							opponentPlayer.Race = "Forfeit"
							userPlayer.Race = "Walkover"
						}

						user.client.Defeat(opponent.client, match_result.Region)
					} else {
						// User loss
						userPlayer.Victory = false
						opponentPlayer.Victory = true

						if result < 53 {
							// User forfeit
							user.client.ForfeitMatchmadeMatch()
							userPlayer.Race = "Forfeit"
							opponentPlayer.Race = "Walkover"
						}

						opponent.client.Defeat(user.client, match_result.Region)
					}

					userPlayer.PointsAfter = userRegion.LadderPoints
					opponentPlayer.PointsAfter = opponentRegion.LadderPoints

					userPlayer.PointsDifference = userPlayer.PointsAfter - userPlayer.PointsBefore
					opponentPlayer.PointsDifference = opponentPlayer.PointsAfter - opponentPlayer.PointsBefore

					// Reset our queue states
					user.client.PendingMatchmakingId = nil
					user.client.PendingMatchmakingOpponentId = nil
					opponent.client.PendingMatchmakingId = nil
					opponent.client.PendingMatchmakingOpponentId = nil

					// Save our work
					err = dbMap.Insert(userPlayer, opponentPlayer)
					if err != nil {
						log.Println(err)
					}
					dbMap.Update(user.client, opponent.client)
				}
			}
		}
	}
}

func doSimulations(count int) {

	// Create all the users
	log.Println("Running simulation with", count, "users")
	users := make([]*SimulatedUser, 0, count)
	for i := 1; i <= count; i++ {
		user := NewSimulatedUser(int64(i), int64(1250), 1)
		users = append(users, user)
	}

	// Now run all the users
	for i := 0; i < len(users); i++ {
		go users[i].Run()
	}

	for {
		time.Sleep(10 * time.Second)
		for i := 0; i < count; i++ {
			log.Printf("%s: %dW %dL, Rating %d", users[i].client.Username, users[i].client.Wins, users[i].client.Losses, users[i].client.LadderPoints)
		}
	}
}
