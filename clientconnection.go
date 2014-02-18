package main

// Connection handler logic

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Starbow/erosd/buffers"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	clientConnections map[int64]*ClientConnection = make(map[int64]*ClientConnection)
	usernameValidator *regexp.Regexp              = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,15}$`)
	connectionIdBase  int64                       = 0
)

const MAXIMUM_DATA_SIZE = 500 * 1024
const READ_BUFFER_SIZE = 4096

type ClientConnection struct {
	id                int64 // Connection ID
	conn              net.Conn
	reader            *bufio.Reader
	writer            *bufio.Writer
	authenticated     bool
	superUser         bool // Maybe allow certain users special functions.
	client            *Client
	logger            *log.Logger
	logFile           *os.File
	lastactive        time.Time
	lastPing          time.Time
	lastPingChallenge string
	latency           int64

	lastPong time.Time

	chatRooms map[string]*ChatRoom // List of rooms we're in.

	sync.RWMutex
}

func NewClientConnection(conn net.Conn) (clientConn *ClientConnection) {
	clientConn = &ClientConnection{
		id:            atomic.AddInt64(&connectionIdBase, 1),
		conn:          conn,
		authenticated: false,
		reader:        bufio.NewReader(conn),
		writer:        bufio.NewWriter(conn),
		chatRooms:     make(map[string]*ChatRoom),
	}
	var source string = conn.RemoteAddr().String()
	var logfile string = path.Join(logPath, fmt.Sprintf("%d-conn-%d.log", os.Getpid(), clientConn.id))
	file, err := os.Create(logfile)
	if err != nil {
		log.Println("Failed to create log file", logfile, "for new connection from", source)
		clientConn.logger = log.New(os.Stdout, fmt.Sprintf("conn-%d:", clientConn.id), log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		log.Println("Logging new connection from", source, "to", logfile)
		clientConn.logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
		clientConn.logFile = file
		clientConn.logger.Println("Logging new connection from", source, "to", logfile)
	}

	clientConnections[clientConn.id] = clientConn
	return clientConn
}

func DisconnectClient(id int64, command string) {
	for _, v := range clientConnections {

		if v.client.Id == id {
			v.SendServerMessage(command, []byte{})
			v.Close()
		}
	}
}

func (conn *ClientConnection) panicRecovery(txid int) {
	if r := recover(); r != nil {
		fmt.Println("Recovered from a panic", r)
		debug.PrintStack()

		conn.SendResponseMessage("106", txid, []byte{})
		// Do we want to disconnect the client here? Might be safer.
	}
}

// Handles sending out messages to everyone

// Client data reader loop goroutine
func (conn *ClientConnection) read() {
	defer conn.panicRecovery(0)
	//Defer executes a function after this function returns.
	defer func() {

		// Handle removing the user from any matchmaking or lobbies they may be in
		delete(clientConnections, conn.id)
		if conn.client != nil {
			delete(clientCharacters, conn.client.Id)
		}
		conn.Close()

		// Dequeue from matchmaking if we're in it.
		el, ok := matchmaker.participants[conn]
		if ok {
			go func() {
				matchmaker.unregister <- conn
				el.abort <- true
			}()
		}

		// Leave all channels.
		for x := range conn.chatRooms {
			conn.chatRooms[x].leave <- conn
		}

		if conn.logFile != nil {
			conn.logFile.Close()
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for _ = range ticker.C {
			if conn.client == nil {
				// We're not authed after 10 seconds. Disconnect.
				conn.Close()
				return
			}

			if conn.lastPingChallenge != "" {
				if time.Since(conn.lastPong).Seconds() > 30 {
					// ping timeout
					conn.Close()
					return
				}
			}

			conn.lastPing = time.Now()
			conn.lastPingChallenge = fmt.Sprintf("%d", conn.lastPing.Second()*conn.lastPing.Minute())
			err := conn.SendServerMessage("PNG", []byte(conn.lastPingChallenge))
			if err != nil {
				return
			}
		}
	}()

	//Infinite loop
	for {
		line, err := conn.reader.ReadString('\n')

		if err != nil {
			// If the error isn't a straight disconnection, print it
			if err != io.EOF {
				conn.logger.Println("Socket Error", err)
			}
			return
		}

		event, txid, length, err := Unpack(line)
		if err != nil {
			return
		}

		if length > MAXIMUM_DATA_SIZE {
			conn.logger.Printf("Connection from %s exceeded max data size (%d)", conn.conn.RemoteAddr().String(), length)
			return
		}

		var data bytes.Buffer

		if length > 0 {

			written, err := io.CopyN(&data, conn.reader, int64(length))
			if err != nil {
				conn.logger.Println(err)
			}

			if written != int64(length) {
				conn.logger.Println("Expecting", length, "got", written)
			}

		}

		conn.Lock()
		conn.lastactive = time.Now()
		conn.Unlock()

		conn.logger.Println("recv:", event, length)

		if conn.client == nil {
			if event == "HSH" {
				if !conn.OnHandshake(txid, data.Bytes()) {
					return
				} else {
					stats := NewServerStats()
					data, err := Marshal(stats)
					if err == nil {
						conn.SendServerMessage("SSU", data)
					}
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
			case "MMF":
				go conn.OnForefeitMatchmaking(txid, data.Bytes())
			case "BNA":
				go conn.OnAddCharacter(txid, data.Bytes())
			case "BNU":
				go conn.OnUpdateCharacter(txid, data.Bytes())
			case "BNR":
				go conn.OnRemoveCharacter(txid, data.Bytes())
			case "REP":
				go conn.OnReplay(txid, data.Bytes())
			//case "UCN":
			//	go conn.OnUserChangeName(txid, data.Bytes())
			case "PNR":
				go conn.OnPong(txid, data.Bytes())
			case "UCJ":
				go conn.OnChatJoin(txid, data.Bytes())
			case "UCL":
				go conn.OnChatLeave(txid, data.Bytes())
			case "UCM":
				go conn.OnChatMessage(txid, data.Bytes())
			case "UPM":
				go conn.OnPrivateMessage(txid, data.Bytes())
			case "UCI":
				go conn.OnChatIndex(txid, data.Bytes())
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
// 204 - Verification failed.

// 301 - Error processing replay
// 302 - Error while processing match result
// 303 - Duplicate Replay
// 304 - The submitting client was not involved in the match.
// 305 - Game too short.
// 306 - Bad format. Required 1v1 with no observers.
// 307 - Bad map. Require a map in the map pool.
// 308 - All participants of the game must be registered.
// 309 - Player not found in database.
// 310 - You didn't play your matchmade opponent. You have been forefeited from that game.
// 401 - Can't queue on this region without a character on this region.
// 402 - The matchmaking request was cancelled.
// 501 - Chat room not joinable.
// 502 - Bad password.
// 503 - Can't create. Already exists.
// 504 - Can't create. Room reserved.
// 505 - Can't join. Max channel limit reached.
// 506 - Can't send message. Not on channel.
// 507 - Can't send message. User offline.
// 508 - Can't send message. Missing fields.
// 509 - Can't create room. Name too short.
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
	} else if err == ErrChatRoomAlreadyExists {
		return "503"
	} else if err == ErrChatRoomReserved {
		return "504"
	} else if err == ErrChatRoomNameTooShort {
		return "509"
	} else {
		return "106"
	}
}

func (conn *ClientConnection) Close() {
	conn.conn.Close()
}

func (conn *ClientConnection) OnChatJoin(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	if int64(len(conn.chatRooms)) >= maxChatRooms {
		conn.SendResponseMessage("505", txid, []byte{})
		return
	}
	var join protobufs.ChatRoomRequest
	err := Unmarshal(data, &join)
	if err != nil {
		conn.Close()
	}

	key := cleanChatRoomName(join.GetRoom())
	room, ok := chatRooms[key]
	if ok {
		if !room.joinable {
			conn.SendResponseMessage("501", txid, []byte{})
			return
		}

		if room.password != "" && room.password != join.GetPassword() {
			conn.SendResponseMessage("502", txid, []byte{})
			return
		}

		info := room.ChatRoomInfoMessage(true)
		data, _ := Marshal(info)

		conn.SendResponseMessage("UCJ", txid, data)

		room.join <- conn
	} else {
		room, err = NewChatRoom(join.GetRoom(), join.GetPassword(), true, false)
		if err != nil {
			conn.SendResponseMessage(ErrorCode(err), txid, []byte(err.Error()))
			return
		}
		room.join <- conn
		info := room.ChatRoomInfoMessage(true)
		data, _ := Marshal(info)

		conn.SendResponseMessage("UCJ", txid, data)
	}
}
func (conn *ClientConnection) OnChatLeave(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	var leave protobufs.ChatRoomRequest
	err := Unmarshal(data, &leave)
	if err != nil {
		conn.logger.Println(err)
		conn.Close()
		return
	}
	key := cleanChatRoomName(leave.GetRoom())
	room, ok := conn.chatRooms[key]
	if ok {
		room.leave <- conn
	}

	conn.SendResponseMessage("UCL", txid, []byte{})

}

func (conn *ClientConnection) OnPrivateMessage(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	var message protobufs.ChatMessage
	err := Unmarshal(data, &message)
	if err != nil {
		conn.Close()
	}

	text := strings.TrimSpace(message.GetMessage())
	target := strings.ToLower(strings.TrimSpace(message.GetTarget()))

	if text == "" || target == "" {
		conn.SendResponseMessage("508", txid, []byte{})
		return
	}

	var outMessage protobufs.ChatPrivateMessage

	stat := conn.client.UserStatsMessage()

	outMessage.Sender = stat
	outMessage.Message = &text
	data, err = Marshal(&outMessage)

	if err != nil {
		conn.SendResponseMessage(ErrorCode(err), txid, []byte{})
		return
	}

	sent := false
	for x := range clientConnections {
		if clientConnections[x].client != nil && strings.ToLower(clientConnections[x].client.Username) == target {
			go clientConnections[x].SendServerMessage("CHP", data)
			sent = true
		}
	}

	if !sent {
		conn.SendResponseMessage("507", txid, []byte{})
	} else {
		conn.logger.Println("Send private message to", message.GetTarget(), ":", message.GetMessage())
		conn.SendResponseMessage("UPM", txid, []byte{})
	}

}

func (conn *ClientConnection) OnChatMessage(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	var message protobufs.ChatMessage
	err := Unmarshal(data, &message)
	if err != nil {
		conn.Close()
	}

	key := cleanChatRoomName(message.GetTarget())
	room, ok := conn.chatRooms[key]
	if ok {
		conn.logger.Println("Sent chat message to room", key, ":", message.GetMessage())
		msg := room.ChatRoomMessageMessage(conn, &message)
		room.message <- &msg
		conn.SendResponseMessage("UCM", txid, []byte{})
	} else {
		conn.SendResponseMessage("506", txid, []byte{})
	}
}
func (conn *ClientConnection) OnChatIndex(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	// This process might be a bit intensive. Perhaps cache it?
	var index protobufs.ChatRoomIndex

	var rooms []*ChatRoom = make([]*ChatRoom, len(joinableChatRooms))

	i := 0
	for x := range joinableChatRooms {

		rooms[i] = joinableChatRooms[x]
		i++
	}

parent:
	for x := range conn.chatRooms {
		for y := range rooms {
			if conn.chatRooms[x] == rooms[y] {
				continue parent
			}
		}

		rooms = append(rooms, conn.chatRooms[x])
	}

	var infos []*protobufs.ChatRoomInfo = make([]*protobufs.ChatRoomInfo, len(rooms))

	i = 0
	for x := range rooms {
		infos[i] = rooms[x].ChatRoomInfoMessage(false)
		i++
	}

	index.Room = infos
	data, err := Marshal(&index)
	if err == nil {
		conn.SendResponseMessage("UCI", txid, data)
	} else {
		conn.SendResponseMessage("106", txid, []byte{})
	}
}

func (conn *ClientConnection) OnPong(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	conn.lastPong = time.Now()
	conn.latency = conn.lastPong.Sub(conn.lastPing).Nanoseconds() / 1000000

	if conn.lastPingChallenge == "" || string(data) != conn.lastPingChallenge {
		conn.Close()
	} else {
		conn.SendResponseMessage("PNR", txid, []byte{})
	}
	conn.lastPingChallenge = ""
}

//TODO: Detmine if we're removing this.
func (conn *ClientConnection) OnUserChangeName(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	username := strings.TrimSpace(string(data))

	if !usernameValidator.MatchString(username) {
		conn.SendResponseMessage("107", txid, []byte{})
		return
	}

	count, _ := dbMap.SelectInt("SELECT COUNT(*) FROM clients WHERE Username=?", username)
	if count > 0 {
		conn.SendResponseMessage("108", txid, []byte{})
		return
	}

	conn.client.Username = username
	dbMap.Update(conn.client)
	conn.SendResponseMessage("UCN", txid, []byte(username))

}

//TODO: Add some sort of real logging at some point
func (conn *ClientConnection) OnReplay(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	file, err := ioutil.TempFile("", "erosreplay")
	if err != nil {
		conn.SendResponseMessage("104", txid, []byte{})
		conn.logger.Println(err)
		return
	}

	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.Write(data)
	if err != nil {
		conn.SendResponseMessage("104", txid, []byte{})
		conn.logger.Println(err)
		return
	}

	file.Close()

	replay, err := NewReplay(file.Name())
	if err != nil {
		conn.SendResponseMessage("301", txid, []byte{})
		conn.logger.Println(err)
		return
	}

	result, players, err := NewMatchResult(replay, conn.client)
	if err != nil {
		conn.SendResponseMessage(ErrorCode(err), txid, []byte(err.Error()))
		conn.logger.Println(err)
		return
	}

	if result == nil {
		conn.SendResponseMessage("301", txid, []byte{})
		return
	}

	message := result.MatchResultMessage(players)

	data, _ = Marshal(message)

	conn.SendResponseMessage("REP", txid, data)
	conn.SendServerMessage("MMI", []byte{})
}

func (conn *ClientConnection) OnAddCharacter(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	if len(data) == 0 {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	region, subregion, id, name := ParseBattleNetProfileUrl(string(data))

	if region == BATTLENET_REGION_UNKNOWN {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	count, err := dbMap.SelectInt("SELECT COUNT(*) FROM battle_net_characters WHERE Region=? and SubRegion=? and ProfileId=?", region, subregion, id)
	if err != nil {
		conn.SendResponseMessage("101", txid, []byte{})
		return
	}

	if count > 0 {
		conn.SendResponseMessage("202", txid, []byte{})
		return
	}

	character := NewBattleNetCharacter(region, subregion, id, name)
	character.ClientId = conn.client.Id
	character.IsVerified = false
	err = character.SetVerificationPortrait()

	if err != nil {
		conn.logger.Println(err)
		conn.SendResponseMessage("203", txid, []byte{})
		return
	}

	err = dbMap.Insert(character)
	if err != nil {
		conn.SendResponseMessage("102", txid, []byte{})
		return
	}

	// This should be its own function
	characterCache.Lock()
	characterCache.characterIds[character.Id] = character
	characterCache.profileIds[character.ProfileIdString()] = character
	characterCache.Unlock()

	payload := character.CharacterMessage()
	data, _ = Marshal(payload)
	conn.SendResponseMessage("BNA", txid, data)
}
func (conn *ClientConnection) OnUpdateCharacter(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	if len(data) == 0 {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	var character_message protobufs.Character

	err := Unmarshal(data, &character_message)

	if err != nil {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	character := characterCache.Get(BattleNetRegion(character_message.GetRegion()), int(character_message.GetSubregion()), int(character_message.GetProfileId()))
	if character == nil {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	if character.ClientId != conn.client.Id {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	updated := false
	if !character.IsVerified {
		ok, err := character.CheckVerificationPortrait()
		if err != nil {
			conn.logger.Println(err)
			conn.SendResponseMessage("203", txid, []byte{})
			return
		}

		if !ok {
			conn.SendResponseMessage("204", txid, []byte{})
			return
		} else {
			character.IsVerified = true
			updated = true
		}
	}

	if character_message.GetIngameProfileLink() != "" && character.InGameProfileLink != character_message.GetIngameProfileLink() {
		if inGameProfileRegex.MatchString(character_message.GetIngameProfileLink()) {
			character.InGameProfileLink = character_message.GetIngameProfileLink()
			updated = true
		}
	}

	if character_message.GetCharacterCode() != 0 && character.CharacterCode != int(character_message.GetCharacterCode()) {
		character.CharacterCode = int(character_message.GetCharacterCode())
		updated = true
	}

	if updated {
		_, err = dbMap.Update(character)
		if err != nil {
			conn.SendResponseMessage("102", txid, []byte{})
			conn.logger.Println(err)
			return
		}
	}

	payload := character.CharacterMessage()
	data, _ = Marshal(payload)
	conn.SendResponseMessage("BNU", txid, data)
}

func (conn *ClientConnection) OnRemoveCharacter(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	if len(data) == 0 {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	var character_message protobufs.Character

	err := Unmarshal(data, &character_message)

	if err != nil {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	character := characterCache.Get(BattleNetRegion(character_message.GetRegion()), int(character_message.GetSubregion()), int(character_message.GetProfileId()))
	if character == nil {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	if character.ClientId != conn.client.Id {
		conn.SendResponseMessage("201", txid, []byte{})
		return
	}

	_, err = dbMap.Delete(character)
	if err != nil {
		conn.SendResponseMessage("102", txid, []byte{})
		conn.logger.Println(err)
		return
	}

	conn.SendResponseMessage("BNR", txid, []byte{})
}

func (conn *ClientConnection) handleMatchmakingResult(txid int, match *MatchmakerMatch, opponent *Client, selectedMap *Map, elapsed int64) {

	var res protobufs.MatchmakingResult
	opponentStats := opponent.UserStatsMessage()
	mapInfo := selectedMap.MapMessage()

	res.Channel = &match.Channel
	res.ChatRoom = &match.ChatRoom
	res.Timespan = &elapsed
	res.Quality = &match.Quality
	res.Opponent = opponentStats
	res.Map = mapInfo

	data, _ := Marshal(&res)
	conn.SendResponseMessage("MMR", txid, data)
	if match.ChatRoom != "" {
		room, ok := chatRooms[cleanChatRoomName(match.ChatRoom)]
		if ok && room != nil {
			room.join <- conn
		}
	}
}

func (conn *ClientConnection) OnQueueMatchmaking(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	_, ok := matchmaker.participants[conn]
	if !ok {

		var queue protobufs.MatchmakingQueue
		err := Unmarshal(data, &queue)
		if err != nil {
			conn.logger.Println(err)
		}

		conn.client.LadderSearchRegion = BattleNetRegion(queue.GetRegion())
		conn.client.LadderSearchRadius = queue.GetRadius()

		if conn.client.LadderSearchRadius < 0 {
			conn.client.LadderSearchRadius = 0
		}

		// Check we have registered characters for this region
		if !conn.client.HasRegion(conn.client.LadderSearchRegion) {
			conn.SendResponseMessage("401", txid, []byte{})
			return
		}

		// Resume pending matches
		if conn.client.PendingMatchmakingId > 0 {
			match := matchmaker.Match(conn.client.PendingMatchmakingId)
			opponent := clientCache.Get(conn.client.PendingMatchmakingOpponentId)

			since := time.Now().Unix() - match.AddTime

			if since >= matchmakingMatchTimeout {
				// Match has expired. End it.
				if opponent != nil {
					matchmaker.EndMatch(conn.client.PendingMatchmakingId, conn.client, opponent)
				} else {
					matchmaker.EndMatch(conn.client.PendingMatchmakingId, conn.client)
				}
			} else {
				// Match is active. Send the old result.
				selectedMap := maps[match.MapId]
				conn.handleMatchmakingResult(txid, match, opponent, selectedMap, 0)
				return
			}
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
				conn.SendResponseMessage("402", txid, data)
				return
			case match := <-el.match:
				opponent := el.opponent
				// We have an opponent! Great success.

				conn.client.PendingMatchmakingId = match.Id
				conn.client.PendingMatchmakingOpponentId = opponent.client.Id
				conn.client.PendingMatchmakingRegion = int64(match.Region)

				dbMap.Update(conn.client)

				conn.handleMatchmakingResult(txid, match, opponent.client, el.selectedMap, int64(time.Since(el.enrollTime).Seconds()))
			}
		}()
		conn.SendResponseMessage("MMQ", txid, []byte{})

	}

}

func (conn *ClientConnection) OnDequeueMatchmaking(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	el, ok := matchmaker.participants[conn]
	if ok {
		go func() {
			matchmaker.unregister <- conn
			el.abort <- true
		}()
	}
	conn.SendResponseMessage("MMD", txid, []byte{})
}

func (conn *ClientConnection) OnForefeitMatchmaking(txid int, data []byte) {
	defer conn.panicRecovery(txid)

	if conn.client.PendingMatchmakingId > 0 {
		conn.client.ForefeitMatchmadeMatch()
	}
	conn.SendResponseMessage("MMF", txid, []byte{})
}

func (conn *ClientConnection) OnSimulation(txid int, data []byte) {
	if !allowsimulations {
		conn.SendResponseMessage("105", txid, []byte{})
		return
	}
	// This is a sqlite query. Wont work elsewhere.
	row, err := dbMap.Select(&Client{}, "SELECT * FROM clients WHERE id != ? ORDER BY RANDOM() LIMIT 1;", conn.client.Id)

	if len(row) == 0 {
		conn.SendResponseMessage("101", txid, []byte{})
		return
	}
	client := (row[0]).(*Client)
	if len(data) == 0 {
		conn.SendResponseMessage("106", txid, []byte{})
		return
	}
	if err == nil {
		var victor int = rand.Intn(2)
		var region BattleNetRegion = BattleNetRegion(data[0])

		if len(data) == 2 {
			if data[1] == 'w' {
				victor = 0
			} else if data[1] == 'l' {
				victor = 1
			}
		}

		var (
			res           protobufs.SimulationResult
			opponentStats *protobufs.UserStats = client.UserStatsMessage()
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
		quality := winner.Defeat(loser, region)

		res.Victory = &victory
		res.Opponent = opponentStats
		res.MatchQuality = &quality

		dbMap.Update(winner, loser)

		data, _ := Marshal(&res)
		conn.SendResponseMessage("SIM", txid, data)

		stats := conn.client.UserStatsMessage()
		data, _ = Marshal(stats)
		conn.SendServerMessage("USU", data)
	}
}

func (conn *ClientConnection) OnHandshake(txid int, data []byte) bool {
	defer conn.panicRecovery(txid)

	var status protobufs.HandshakeResponse_HandshakeStatus = protobufs.HandshakeResponse_FAIL
	var resp protobufs.HandshakeResponse
	defer func() {

		resp.Status = &status

		data, err := Marshal(&resp)
		if err == nil {
			conn.SendResponseMessage("HSH", txid, data)
		}
	}()

	var hs protobufs.Handshake
	err := Unmarshal(data, &hs)
	if err != nil {
		conn.logger.Println(err)
		return false
	}

	var client *Client

	realUser := GetRealUser(hs.GetUsername(), hs.GetAuthKey())

	if realUser == nil {
		conn.logger.Println("bad auth", hs.GetUsername(), hs.GetAuthKey())
		return false
	}

	client = clientCache.Get(realUser.Id)

	if client == nil {
		client = NewClient(realUser.Id)

		err := dbMap.Insert(client)
		clientCache.clients[client.Id] = client
		client.Username = realUser.Username

		dbMap.Update(client)

		conn.logger.Printf("New client %+v %+v", *client, err)
	}

	conn.logger.Printf("Client %+v", *client)
	conn.client = client

	var user *protobufs.UserStats = client.UserStatsMessage()

	characters, err := client.Characters()

	if err == nil {
		var characterMessages []*protobufs.Character = make([]*protobufs.Character, len(characters))

		for x := range characters {
			characterMessages[x] = characters[x].CharacterMessage()
		}

		resp.Character = characterMessages
	}

	resp.User = user
	resp.Id = &client.Id
	resp.Division = make([]*protobufs.Division, 0, len(divisions))
	resp.ActiveRegion = make([]protobufs.Region, 0, len(ladderActiveRegions))
	for x := range divisions {
		resp.Division = append(resp.Division, divisions[x].DivisionMessage())
	}
	for _, region := range ladderActiveRegions {
		resp.ActiveRegion = append(resp.ActiveRegion, protobufs.Region(region))
	}
	status = protobufs.HandshakeResponse_SUCCESS

	return true
}

func (conn *ClientConnection) SendResponseMessage(command string, txid int, data []byte) error {

	header := fmt.Sprintf("%s %d %d", command, txid, len(data))
	conn.logger.Println("send:", header)
	conn.Lock()
	defer conn.Unlock()
	_, err := conn.writer.WriteString(header)

	if err != nil {
		return err
	}

	err = conn.writer.WriteByte('\n')
	if err != nil {
		return err
	}
	_, err = conn.writer.Write(data)
	if err != nil {
		return err
	}
	err = conn.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (conn *ClientConnection) SendServerMessage(command string, data []byte) error {

	header := fmt.Sprintf("%s %d", command, len(data))
	conn.logger.Println("send:", header)
	conn.Lock()
	defer conn.Unlock()
	_, err := conn.writer.WriteString(header)
	if err != nil {
		return err
	}
	err = conn.writer.WriteByte('\n')
	if err != nil {
		return err
	}
	_, err = conn.writer.Write(data)
	if err != nil {
		return err
	}
	err = conn.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
