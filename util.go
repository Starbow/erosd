package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"github.com/Starbow/erosd/buffers"
	"log"
	"strconv"
	"strings"
)

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
	log.Println(data, event, size)
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
	data, err := Marshal(message)

	if err != nil {
		panic(err)
	}
	for _, v := range clientConnections {

		go v.SendServerMessage(command, data)
	}
}

//Broadcast a message to a specific client.
func (c *Client) Broadcast(command string, message proto.Message) {
	data, err := Marshal(message)

	if err != nil {
		panic(err)
	}
	for _, v := range clientConnections {

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