package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
)

var _ = log.Ldate
var (
	pythonPath string
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
	cmd := exec.Command(pythonPath, "sc2json.py", "--indent", "4", path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	var rep Replay
	data, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &rep)
	if err != nil {
		return
	}
	replay = &rep
	return
}
