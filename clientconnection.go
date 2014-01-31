package main

// Connection handler logic

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"github.com/Starbow/erosd/buffers"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	clientConnections *list.List     = list.New()
	usernameValidator *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,15}$`)
)

const MAXIMUM_DATA_SIZE = 500 * 1024
const READ_BUFFER_SIZE = 4096

type ClientConnection struct {
	conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	authenticated bool
	superUser     bool // Maybe allow certain users special functions.
	client        *Client

	lastactive time.Time
	lastPing   time.Time
	lastPong   time.Time

	sync.RWMutex
}

func NewClientConnection(conn net.Conn) (clientConn *ClientConnection) {
	clientConn = &ClientConnection{
		conn:          conn,
		authenticated: false,
		reader:        bufio.NewReader(conn),
		writer:        bufio.NewWriter(conn),
	}

	clientConnections.PushBack(clientConn)
	return clientConn
}

func DisconnectClient(id int64, command string) {
	for e := clientConnections.Front(); e != nil; e = e.Next() {
		client := e.Value.(*ClientConnection)
		if client.client.Id == id {
			client.SendData(command, 0, []byte{})
			client.conn.Close()
		}
	}
}

// Generates server stats protocol buffer message. This should be elsewhere maybe.

func NewServerStats() protobufs.ServerStats {
	var (
		x         protobufs.ServerStats
		connected int64 = int64(clientConnections.Len())
		mm        int64 = int64(len(matchmaker.participants))
	)
	x.ActiveUsers = &connected
	x.MatchmakingUsers = &mm

	return x
}

// Handles sending out messages to everyone

// Client data reader loop goroutine
func (conn *ClientConnection) read() {
	//Defer executes a function after this function returns.
	defer func() {
		// Handle removing the user from any matchmaking or lobbies they may be in
		for e := clientConnections.Front(); e != nil; e = e.Next() {
			if e.Value == conn {
				clientConnections.Remove(e)
				break
			}
		}
		conn.conn.Close()

		el, ok := matchmaker.participants[conn]
		if ok {
			go func() {
				matchmaker.unregister <- conn
				el.abort <- true
			}()
		}
	}()

	//Infinite loop
	for {
		line, err := conn.reader.ReadString('\n')

		if err != nil {
			// If the error isn't a straight disconnection, print it
			if err != io.EOF {
				log.Println("Socket Error", err)
			}
			return
		}

		event, txid, length, err := Unpack(line)
		if err != nil {
			return
		}

		if length > MAXIMUM_DATA_SIZE {
			log.Printf("Connection from %s exceeded max data size (%d)", conn.conn.RemoteAddr().String(), length)
			return
		}

		//data := make([]byte, 0, length)
		var data bytes.Buffer

		if length > 0 {

			written, err := io.CopyN(&data, conn.reader, int64(length))
			if err != nil {
				log.Println(err)
			}

			if written != int64(length) {
				log.Println("Expecting", length, "got", written)
			}

		}

		conn.Lock()
		conn.lastactive = time.Now()
		conn.Unlock()

		if conn.client == nil {
			if event == "HSH" {
				if !conn.OnHandshake(txid, data.Bytes()) {
					return
				}
			} else {
				return
			}
		} else {
			// dispatch
			switch event {
			case "SIM":
				go conn.OnSimulation(txid, data.Bytes())
			case "MMQ":
				go conn.OnQueueMatchmaking(txid, data.Bytes())
			case "MMD":
				go conn.OnDequeueMatchmaking(txid, data.Bytes())
			case "CHA":
				go conn.OnAddCharacter(txid, data.Bytes())
			case "CHV":
				go conn.OnVerifyCharacter(txid, data.Bytes())
			case "REP":
				go conn.OnReplay(txid, data.Bytes())
			case "UCN":
				go conn.OnUserChangeName(txid, data.Bytes())
			}
		}
	}
}

// Error codes:
// 1xx - Internal Server Errors
// 2xx - Battle.net Errors
// 3xx - Ladder errors
// 4xx - Matchmaking error

// 101 - Database read error
// 102 - Database write error
// 103 - Disk read error
// 104 - Disk write error
// 105 - Authentication error
// 106 - Generic error
// 107 - Bad name.
// 108 - Name in use.
// 201 - Bad character info
// 202 - Character already exists
// 203 - Error while communicating with Battle.net

// 301 - Error processing replay
// 302 - Error while processing match result
// 303 - Duplicate Replay
// 303 - A player was not found in the database.
// 304 - The submitting client was not involved in the match.
// 305 - Game too short.
// 306 - Bad format. Required 1v1 with no observers.
// 307 - Bad map. Require a map in the map pool.
// 308 - All participants of the game must be registered.
// 309 - Player not found in database.
// 310 - You didn't play your matchmade opponent. You have been forefeited from that game.
// 401 - Can't queue on this region without a character on this region.

func ErrorCode(err error) string {
	if err == ErrLadderClientNotInvolved {
		return "304"
	} else if err == ErrLadderDuplicateReplay {
		return "303"
	} else if err == ErrDbInsert {
		return "102"
	} else if err == ErrLadderGameTooShort {
		return "305"
	} else if err == ErrLadderInvalidFormat {
		return "306"
	} else if err == ErrLadderInvalidMap {
		return "307"
	} else if err == ErrLadderInvalidMatchParticipents {
		return "308"
	} else if err == ErrLadderPlayerNotFound {
		return "309"
	} else if err == ErrLadderWrongOpponent {
		return "310"
	} else {
		return "106"
	}
}

func (conn *ClientConnection) OnUserChangeName(txid int, data []byte) {
	username := strings.TrimSpace(string(data))

	if !usernameValidator.MatchString(username) {
		conn.SendData("107", txid, []byte{})
		return
	}

	count, _ := dbMap.SelectInt("SELECT COUNT(*) FROM clients WHERE Username=?", username)
	if count > 0 {
		conn.SendData("108", txid, []byte{})
		return
	}

	conn.client.Username = username
	dbMap.Update(conn.client)
	conn.SendData("UCN", txid, []byte(username))

}

//Add some sort of real logging at some point
func (conn *ClientConnection) OnReplay(txid int, data []byte) {
	file, err := ioutil.TempFile("", "erosreplay")
	if err != nil {
		conn.SendData("104", txid, []byte{})
		log.Println(err)
		return
	}

	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.Write(data)
	if err != nil {
		conn.SendData("104", txid, []byte{})
		log.Println(err)
		return
	}

	file.Close()

	replay, err := NewReplay(file.Name())
	if err != nil {
		conn.SendData("301", txid, []byte{})
		log.Println(err)
		return
	}

	result, players, err := NewMatchResult(replay, conn.client)
	if err != nil {
		conn.SendData(ErrorCode(err), txid, []byte(err.Error()))
		log.Println(err)
		return
	}

	if result != nil {
		log.Printf("%+v", *result)
		log.Printf("%+v", *players[0])
		log.Printf("%+v", *players[1])
	}
}

func (conn *ClientConnection) OnAddCharacter(txid int, data []byte) {
	if len(data) == 0 {
		conn.SendData("201", txid, []byte{})
		return
	}

	region, subregion, id, name := ParseBattleNetProfileUrl(string(data))

	if region == BATTLENET_REGION_UNKNOWN {
		conn.SendData("201", txid, []byte{})
		return
	}

	count, err := dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE Region=? and SubRegion=? and ProfileId=?", region, subregion, id)
	if err != nil {
		conn.SendData("101", txid, []byte{})
		return
	}

	if count > 0 {
		//conn.SendData("202", txid, []byte{})
		//return
	}

	character := NewBattleNetCharacter(region, subregion, id, name)
	character.ClientId = conn.client.Id
	character.IsVerified = true
	err = character.SetVerificationPortrait()

	if err != nil {
		log.Println(err)
		conn.SendData("203", txid, []byte{})
		return
	}

	err = dbMap.Insert(character)
	if err != nil {
		conn.SendData("102", txid, []byte{})
		return
	}

	// This should be its own function
	characterCache.Lock()
	characterCache.characterIds[character.Id] = character
	characterCache.profileIds[character.ProfileIdString()] = character
	characterCache.Unlock()

	payload := character.CharacterMessage()
	data, _ = Marshal(&payload)
	conn.SendData("CHA", txid, data)
}
func (conn *ClientConnection) OnVerifyCharacter(txid int, data []byte) {
}

func (conn *ClientConnection) OnQueueMatchmaking(txid int, data []byte) {
	_, ok := matchmaker.participants[conn]
	if !ok {

		var queue protobufs.MatchmakingQueue
		err := Unmarshal(data, &queue)
		if err != nil {
			log.Println("wat", err)
		}

		conn.client.LadderSearchRegion = BattleNetRegion(queue.GetRegion())
		conn.client.LadderSearchRadius = queue.GetRadius()

		if conn.client.LadderSearchRadius < 0 {
			conn.client.LadderSearchRadius = 0
		}

		// Check we have registered characters for this region
		if !conn.client.HasRegion(conn.client.LadderSearchRegion) {
			conn.SendData("401", txid, []byte{})
			return
		}

		matchmaker.register <- conn
		//We need to wait until the matchmaker has actually finished adding
		//our new participant, otherwise our lookup would fail.
		<-matchmaker.callback
		el := matchmaker.participants[conn]
		go func() {
			select {
			case <-el.abort:
				// We've been cancelled somewhere. Abort out.
				return
			case match := <-el.match:
				opponent := el.opponent
				// We have an opponent! Great success.

				conn.client.PendingMatchmakingId = match.Id
				conn.client.PendingMatchmakingOpponentId = opponent.client.Id

				dbMap.Update(conn.client)

				var res protobufs.MatchmakingResult

				elapsed := int64(time.Since(el.enrollTime).Seconds())
				opponentStats := opponent.client.UserStatsMessage()
				mapInfo := el.selectedMap.MapMessage()

				res.Channel = &match.Channel
				res.Timespan = &elapsed
				res.Quality = &match.Quality
				res.Opponent = &opponentStats
				res.Map = &mapInfo

				data, _ := Marshal(&res)
				conn.SendData("MMR", txid, data)
			}
		}()
		conn.SendData("MMQ", txid, []byte{})

	}

}

func (conn *ClientConnection) OnDequeueMatchmaking(txid int, data []byte) {
	el, ok := matchmaker.participants[conn]
	if ok {
		go func() {
			matchmaker.unregister <- conn
			el.abort <- true
		}()
	}
	conn.SendData("MMD", txid, []byte{})
}

func (conn *ClientConnection) OnSimulation(txid int, data []byte) {

	row, err := dbMap.Select(&Client{}, "SELECT * FROM clients WHERE id != ? ORDER BY RANDOM() LIMIT 1;", conn.client.Id)

	if len(row) == 0 {
		return
	}
	client := (row[0]).(*Client)

	if err == nil {
		var victor int = rand.Intn(2)
		if len(data) == 1 {
			if data[0] == 'w' {
				victor = 0
			} else if data[0] == 'l' {
				victor = 1
			}
		}

		var (
			res           protobufs.SimulationResult
			opponentStats protobufs.UserStats = client.UserStatsMessage()
			winner        *Client
			loser         *Client
			victory       bool
		)

		if victor == 0 {
			winner = conn.client
			loser = client
			victory = true
		} else {
			winner = client
			loser = conn.client
			victory = false
		}
		quality := winner.Defeat(loser)

		res.Victory = &victory
		res.Opponent = &opponentStats
		res.MatchQuality = &quality

		dbMap.Update(winner, loser)

		data, _ := Marshal(&res)
		conn.SendData("SIM", txid, data)

		stats := conn.client.UserStatsMessage()
		data, _ = Marshal(&stats)
		conn.SendData("USU", txid, data)
	}
}

func (conn *ClientConnection) OnHandshake(txid int, data []byte) bool {
	var hs protobufs.Handshake
	log.Println(string(data))
	err := Unmarshal(data, &hs)
	if err != nil {
		log.Println("wat", err)
	}

	var client *Client
	var resp protobufs.HandshakeResponse
	var status protobufs.HandshakeResponse_HandshakeStatus = protobufs.HandshakeResponse_SUCCESS
	resp.Status = &status

	if hs.GetId() == 0 {
		client = NewClient()

		err := dbMap.Insert(client)
		clientCache.clients[client.Id] = client
		client.Username = fmt.Sprintf("Anonymous%d", client.Id)
		dbMap.Update(client)

		log.Printf("New client %+v %+v", *client, err)
	} else {
		client = clientCache.Get(hs.GetId())
		if client == nil {
			log.Println("Client not found")
			return false
		}
		auth := hs.GetAuthKey()
		if client.AuthKey != auth {
			log.Println("bad uuth", client.AuthKey, auth)
			return false
		}
	}
	log.Printf("Client %+v", *client)
	conn.client = client

	var user protobufs.UserStats = client.UserStatsMessage()

	resp.User = &user
	resp.Id = &client.Id
	resp.AuthKey = &client.AuthKey

	data, err = Marshal(&resp)
	conn.SendData("HSH", txid, data)
	stats := NewServerStats()
	data, err = Marshal(&stats)
	conn.SendData("SSU", txid, data)

	return true
}

func (conn *ClientConnection) SendData(command string, txid int, data []byte) {
	header := fmt.Sprintf("%s %d %d\n", command, txid, len(data))
	log.Println("Send", header)
	conn.Lock()
	conn.writer.WriteString(header)
	conn.writer.Write(data)
	conn.writer.Flush()
	conn.Unlock()
}
