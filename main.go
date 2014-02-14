package main

import (
	conf "github.com/msbranco/goconfig"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

var (
	randomSource     rand.Source = rand.NewSource(time.Now().Unix())
	listen           string
	simulator        bool
	allowsimulations bool
	matchmaker       *Matchmaker
	testMode         bool
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

// Handle an incoming connection.
func handleConnection(conn net.Conn) {
	client := NewClientConnection(conn)
	client.read()
}

func broadcastRunner() {
	// This broadcasts server stats to the masses every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for _ = range ticker.C {
			x := NewServerStats()
			broadcastMessage("SSU", x)
		}
	}()
}

func doSimulations(count int) {
	i := 0
	points := int64(1250)
	var sims []*ClientSimulation = make([]*ClientSimulation, 0, count)
	for {
		if i >= count {
			break
		}

		go func() {
			sim := NewClientSimulation(points, 1)
			points += 50
			sims = append(sims, sim)
			log.Println("Launching sim", sim.client.Username)
			sim.Run()
		}()

		i++
	}

	for {
		time.Sleep(10 * time.Second)
		go func() {
			i := 0
			for {
				if i >= count {
					break
				}

				//log.Printf("%s: %dW %dL, Rating %d, Avg Queue %f", sims[i].client.Username, sims[i].client.Wins, sims[i].client.Losses, sims[i].client.LadderPoints, sims[i].client.TotalQueueTime/float64(sims[i].client.Wins+sims[i].client.Losses))

				i++
			}
		}()
	}
}

func loadConfig() error {
	config, err := conf.ReadConfigFile("erosd.cfg")
	if err != nil {
		config = conf.NewConfigFile()
		config.AddSection("erosd")
		config.AddOption("erosd", "listen", ":12345")
		config.AddOption("erosd", "simulator", "false")
		config.AddOption("erosd", "allowsimulations", "false")
		config.AddOption("erosd", "python", "/usr/bin/python2.7")
		config.AddOption("erosd", "testmode", "false")

		config.AddSection("ladderdivisions")
		config.AddOption("ladderdivisions", "divisions", "4")
		config.AddOption("ladderdivisions", "subdivisions", "2")
		config.AddOption("ladderdivisions", "subdivisionpoints", "1000")
		config.AddOption("ladderdivisions", "divisionnames", "Bronze;Silver;Gold;Platinum;Diamond;Master;Grand Master")

		config.AddSection("ladder")
		config.AddOption("ladder", "startingpoints", "1250")
		config.AddOption("ladder", "winpointsbase", "100")
		config.AddOption("ladder", "winpointsincrement", "25")
		config.AddOption("ladder", "losepointsbase", "50")
		config.AddOption("ladder", "losepointsincrement", "12.5")
		config.AddOption("ladder", "maxvetos", "3")

		config.AddSection("database")
		config.AddOption("database", "type", "sqlite3")
		config.AddOption("database", "connection", "erosd.sqlite3")

		config.AddSection("chat")
		config.AddOption("chat", "fixedrooms", "Practice Partner Search (Bronze-Silver);Practice Partner Search (Gold-Platinum);Practice Partner Search")
		config.AddOption("chat", "maxuserchats", "5")

		err = config.WriteConfigFile("erosd.cfg", 0644, "Erosd Config")
		if err != nil {
			return err
		}
	}

	listen, _ = config.GetString("erosd", "listen")
	simulator, _ = config.GetBool("erosd", "simulator")
	pythonPath, _ = config.GetString("erosd", "python")
	testMode, _ = config.GetBool("erosd", "testmode")
	allowsimulations, _ = config.GetBool("erosd", "allowsimulations")

	divisionCount, _ = config.GetInt64("ladderdivisions", "divisions")
	subdivisionCount, _ = config.GetInt64("ladderdivisions", "subdivisions")
	divisionPoints, _ = config.GetInt64("ladderdivisions", "subdivisionpoints")
	dn, err := config.GetString("ladderdivisions", "divisionnames")
	if err == nil {
		divisionNames = strings.Split(dn, ";")
	}

	ladderStartingPoints, _ = config.GetInt64("ladder", "startingpoints")
	ladderWinPointsBase, _ = config.GetInt64("ladder", "winpointsbase")
	ladderWinPointsIncrement, _ = config.GetFloat("ladder", "winpointsincrement")
	ladderLosePointsBase, _ = config.GetInt64("ladder", "losepointsbase")
	ladderLosePointsIncrement, _ = config.GetFloat("ladder", "losepointsincrement")
	ladderMaxMapVetos, _ = config.GetInt64("ladder", "maxvetos")
	rg, err := config.GetString("ladder", "activeregions")
	if err == nil {
		regions := strings.Split(rg, ";")
		ladderActiveRegions = make([]BattleNetRegion, 0, len(regions))

		for _, region := range regions {
			regionCode := ParseBattleNetRegion(region)
			if regionCode != BATTLENET_REGION_UNKNOWN {
				ladderActiveRegions = append(ladderActiveRegions, regionCode)
			}
		}
	}

	dbType, _ = config.GetString("database", "type")
	dbConnectionString, _ = config.GetString("database", "connection")

	cn, err := config.GetString("chat", "fixedrooms")
	if err == nil {
		fixedChatRooms = strings.Split(cn, ";")
	}
	maxChatRooms, _ = config.GetInt64("chat", "maxuserchats")
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := loadConfig()
	if err != nil {
		log.Panicln("Error while loading config", err)
	}

	if simulator {
		go matchmaker.run()
		doSimulations(25)
		return
	}

	err = initDb()
	if err != nil {
		log.Fatalln("initDb", err)
	}

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalln("Failed to listen on", listen, err)
	}

	// start the broadcast routine
	go broadcastRunner()

	initDivisions()
	initClientCaches()
	initBattleNet()
	initMatchmaking()
	initChat()

	// start the matchmaker

	//Accept connections forever.
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error", err)
			continue
		}
		log.Println("Accepted", conn.RemoteAddr())

		go handleConnection(conn)

	}

}
