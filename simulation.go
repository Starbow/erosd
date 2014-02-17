package main

import (
    "log"
    "math/rand"
    "time"
    "fmt"
    "github.com/Starbow/erosd/buffers"
)


type SimulatedUser struct {
    client *Client
    conn   *ClientConnection
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
        dbMap.Insert(client)
    }

    // Create a character for this client with a random region
    region := region_choices[rand.Intn(len(region_choices))]
    character := characterCache.Get(region, 1, int(site_user.Id))
    if character == nil {
        log.Println("Creating new character for user", username, "in region", region)
        character = NewBattleNetCharacter(region, 1, int(site_user.Id), site_user.Username)
        character.IsVerified = true
        character.ClientId = client.Id
        dbMap.Insert(character)
    }

    // Create a dummy connection for the client
    connection := &ClientConnection{
        id: site_user.Id,
        conn: nil,
        client: client,
        authenticated: true,
    }

    // Return a fully packed SimulatedUser
    return &SimulatedUser{
        client: client,
        conn: connection,
        characters: []*BattleNetCharacter{character},
    }
}

func (user *SimulatedUser) Run() {
    races := []string{"Terran", "Zerg", "Protoss"}
    for {
        // Pick a random character to queue with
        character := user.characters[rand.Intn(len(user.characters))]
        user.client.LadderSearchRegion = character.Region

        // Join the matchmaking queue
        matchmaker.register <- user.conn

        // Wait until we have been queued
        <- matchmaker.callback

        el, ok := matchmaker.participants[user.conn]
        if ok {
            // Wait until a match is found
            match := <-el.match
            opponent := el.opponent

            // Update the database record for the client
            user.client.PendingMatchmakingId = match.Id
            user.client.PendingMatchmakingOpponentId = opponent.client.Id
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

                // Insert a match result record for this match
                bnet_map := maps.Get(character.Region, *mm_result.Map.BattleNetName)

                match_result := &MatchResult{
                    DateTime: time.Now().Unix(),
                    MapId: bnet_map.Id,
                    MatchmakerMatchId: match.Id,
                    Region: user.client.LadderSearchRegion,
                }
                dbMap.Insert(match_result);

                // Get region stats for us and them
                userRegion, _ := user.client.RegionStats(character.Region)
                opponentRegion, _ := opponent.client.RegionStats(character.Region)

                // Create player record for us
                userPlayer := &MatchResultPlayer{
                    MatchId: match.Id,
                    ClientId: user.client.Id,
                    CharacterId: character.Id,
                    Race: races[rand.Intn(len(races))],
                }

                // Create player record for them
                // TODO: Get the other player's characterId
                opponentPlayer := &MatchResultPlayer{
                    MatchId: match.Id,
                    ClientId: opponent.client.Id,
                    CharacterId: 0,
                    Race: races[rand.Intn(len(races))],
                }

                userPlayer.PointsBefore = userRegion.LadderPoints
                opponentPlayer.PointsBefore = opponentRegion.LadderPoints

                // Assign results for each player
                result := rand.Intn(100)
                if result < 50 {
                    // User win
                    if result < 3 {
                        // User walkover
                        opponent.client.ForefeitMatchmadeMatch()
                    }
                    user.client.Defeat(opponent.client, match_result.Region)
                } else {
                    // User loss
                    if result < 53 {
                        // User forefeit
                        user.client.ForefeitMatchmadeMatch()
                    }
                    opponent.client.Defeat(user.client, match_result.Region)
                }

                userPlayer.PointsAfter = userRegion.LadderPoints
                opponentPlayer.PointsAfter = opponentRegion.LadderPoints

                // Reset our queue states
                user.client.PendingMatchmakingId = 0
                user.client.PendingMatchmakingOpponentId = 0
                opponent.client.PendingMatchmakingId = 0
                opponent.client.PendingMatchmakingOpponentId = 0

                // Save our work
                dbMap.Insert(userPlayer, opponentPlayer)
                dbMap.Update(user.client, opponent.client)
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
