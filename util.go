package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/Starbow/erosd/buffers"
	"runtime/debug"
	"strconv"
	"strings"
)

func genericPanicRecover() {
	if r := recover(); r != nil {
		fmt.Println("Recovered from a panic", r)
		debug.PrintStack()
	}
}

//Take a "CMD TxID Len\n" input and split it up
func Unpack(data string) (event string, txid int, size int, err error) {
	data = strings.TrimRight(data, "\n")
	data = strings.TrimRight(data, "\r")
	result := strings.Split(data, " ")
	if len(result) != 3 {
		err = errors.New("Unable to extract event data.")
		return
	}
	txid, err = strconv.Atoi(result[1])
	if err != nil {
		return
	}

	size, err = strconv.Atoi(result[2])
	if err != nil {
		return
	}

	event = result[0]
	err = nil
	return
}

//Data -> proto.Message
func Unmarshal(data []byte, message proto.Message) error {
	return proto.Unmarshal(data, message)
}

//message -> data
func Marshal(message proto.Message) (data []byte, err error) {
	return proto.Marshal(message)
}

//Broadcast a message to all active connections.
func broadcastMessage(command string, message proto.Message) {
	defer genericPanicRecover()
	data, err := Marshal(message)

	if err != nil {
		panic(err)
	}
	for _, v := range clientConnections {
		if v == nil {
			continue
		}

		if v.client == nil {
			continue
		}
		go v.SendServerMessage(command, data)
	}
}

//Broadcast a message to a specific client.
func (c *Client) Broadcast(command string, message proto.Message) {
	defer genericPanicRecover()
	var (
		data []byte = []byte{}
		err  error
	)
	if message != nil {
		data, err = Marshal(message)

		if err != nil {
			panic(err)
		}
	}
	for _, v := range clientConnections {
		if v == nil {
			continue
		}
		if v.client.Id == c.Id {
			go v.SendServerMessage(command, data)
		}
	}
}

// Generates server stats protocol buffer message. This should be elsewhere maybe.

func NewMatchmakingStats(region BattleNetRegion) *protobufs.MatchmakingStats {
	var (
		stats       protobufs.MatchmakingStats
		protoRegion protobufs.Region = protobufs.Region(region)
		searching   int64            = 0
	)
	if _, ok := matchmaker.regionParticipants[region]; ok {
		searching = int64(len(matchmaker.regionParticipants[region]))
	}

	stats.Region = &protoRegion
	stats.SearchingUsers = &searching

	return &stats
}

func NewServerStats() *protobufs.ServerStats {
	var (
		x         protobufs.ServerStats
		connected int64 = int64(len(clientConnections))
		mm        int64 = int64(len(matchmaker.participants))
	)

	x.ActiveUsers = &connected
	x.SearchingUsers = &mm
	x.Region = []*protobufs.MatchmakingStats{
		NewMatchmakingStats(BATTLENET_REGION_NA),
		NewMatchmakingStats(BATTLENET_REGION_EU),
		NewMatchmakingStats(BATTLENET_REGION_KR),
		NewMatchmakingStats(BATTLENET_REGION_CN),
		NewMatchmakingStats(BATTLENET_REGION_SEA),
	}

	return &x
}

func SendBroadcastAlert(predefined int32, message string) {
	var bufmsg protobufs.BroadcastAlert
	bufmsg.Message = &message
	bufmsg.Predefined = &predefined

	broadcastMessage("ALT", &bufmsg)
}

func ErosErrors(error_code int) error {
	available_errors := map[int]string{
		101: "Database read error",
		102: "Database write error",
		103: "Disk read error",
		104: "Disk write error",
		105: "Authentication error",
		106: "Generic error",
		107: "Bad name.",
		108: "Name in use.",
		109: "Cannot do that while matched via matchmaking.",
		201: "Bad character info",
		202: "Character already exists",
		203: "Error while communicating with Battle.net",
		204: "Verification failed.",

		301: "Error processing replay",
		302: "Error while processing match result",
		303: "Duplicate Replay",
		304: "The submitting client was not involved in the match.",
		305: "Game too short.",
		306: "Bad format. Required 1v1 with no observers.",
		307: "Bad map. Require a map in the map pool.",
		308: "All participants of the game must be registered.",
		309: "Player not found in database.",
		310: "You didn't play your matchmade opponent. You have been forfeited from that game.",
		311: "The game was not played on Faster.",
		312: "Cannot add veto. Map not in ranked pool.",
		313: "Cannot add veto. Maximum number of vetoes used.",
		314: "You are not in a game arranged by the Eros matchmaker.",
		401: "Can't queue on this region without a character on this region.",
		402: "The matchmaking request was cancelled.",
		403: "Long process unavailable.",
		501: "Chat room not joinable.",
		502: "Bad password.",
		503: "Can't create. Already exists.",
		504: "Can't create. Room reserved.",
		505: "Can't join. Max channel limit reached.",
		506: "Can't send message. Not on channel.",
		507: "Can't send message. User offline.",
		508: "Can't send message. Missing fields.",
		509: "Can't create room. Name too short.",
		510: "Can't send message. Rate limit.",
		511: "Can't send message. Message too long.",
	}

	return errors.New(available_errors[error_code])
}
