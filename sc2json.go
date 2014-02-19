package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
)

var _ = log.Ldate
var (
	pythonPort string
)

type ReplayPlayer struct {
	Uid       int64  `json:"uid"`
	GameRace  string `json:"play_race"`
	LobbyRace string `json:"pick_race"`
	Handicap  int64  `json:"handicap"`
	Pid       int    `json:"pid"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Victory   string `json:"result"`
}
type Replay struct {
	IsLadder      bool           `json:"is_ladder"`
	RealType      string         `json:"real_type"`
	Frames        int64          `json:"frames"`
	Speed         string         `json:"speed"`
	Category      string         `json:"category"`
	Filehash      string         `json:"filehash"`
	Build         int64          `json:"build"`
	GameLength    int64          `json:"game_length"`
	RealLength    int64          `json:"real_length"`
	GameFps       float64        `json:"game_fps"`
	UnixTimestamp int64          `json:"unix_timestamp"`
	Type          string         `json:"type"`
	Observers     []ReplayPlayer `json:"observers"`
	Players       []ReplayPlayer `json:"players"`
	MapName       string         `json:"map_name"`
	Versions      []int          `json:"versions"`
	Region        string         `json:"region"`
	Release       string         `json:"release"`
}

func NewReplay(path string) (replay *Replay, err error) {
	conn, err := net.Dial("tcp", pythonPort)
	if err != nil {
		matchmaker.logger.Println("Dial error", err)
		return nil, err
	}

	defer conn.Close()

	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(path + "\n")
	if err != nil {
		return nil, err
	}

	err = writer.Flush()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			log.Println("Socket Error", err)
		}
		return
	}

	if len(data) == 0 {
		matchmaker.logger.Println("Data retreieved for", path, "empty")
		return nil, nil
	}

	var rep Replay
	err = json.Unmarshal(data, &rep)
	if err != nil {
		return nil, err
	}

	replay = &rep
	return
}
