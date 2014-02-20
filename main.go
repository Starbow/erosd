package main

import (
	conf "github.com/msbranco/goconfig"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	randomSource         rand.Source = rand.NewSource(time.Now().Unix())
	listenAddresses      []string
	adminListenAddresses []string
	simulator            bool
	allowsimulations     bool
	matchmaker           *Matchmaker
	testMode             bool
	logPath              string
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

// Handle an incoming connection.
func handleConnection(conn net.Conn) {
	client := NewClientConnection(conn)
	client.read()
}

func handleAdminConnection(conn net.Conn) {
	client := NewAdminConnection(conn)
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

func loadConfig() error {
	config, err := conf.ReadConfigFile("erosd.cfg")
	if err != nil {
		config = conf.NewConfigFile()
		config.AddSection("erosd")
		config.AddOption("erosd", "listen", ":12345")
		config.AddOption("erosd", "adminlisten", "127.0.0.1:12346")
		config.AddOption("erosd", "simulator", "false")
		config.AddOption("erosd", "allowsimulations", "false")
		config.AddOption("erosd", "testmode", "false")
		config.AddOption("erosd", "logpath", "logs")

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
		config.AddOption("ladder", "matchtimeout", "7200")

		config.AddSection("database")
		config.AddOption("database", "type", "sqlite3")
		config.AddOption("database", "connection", "erosd.sqlite3")

		config.AddSection("chat")
		config.AddOption("chat", "fixedrooms", "Practice Partner Search (Bronze-Silver);Practice Partner Search (Gold-Platinum);Practice Partner Search")
		config.AddOption("chat", "maxuserchats", "5")

		config.AddSection("python")
		config.AddOption("python", "port", ":54321")

		err = config.WriteConfigFile("erosd.cfg", 0644, "Erosd Config")
		if err != nil {
			return err
		}
	}

	listen, err := config.GetString("erosd", "listen")
	if err == nil {
		listenAddresses = strings.Split(listen, ";")
	}
	listen, err = config.GetString("erosd", "adminlisten")
	if err == nil {
		adminListenAddresses = strings.Split(listen, ";")
	}
	simulator, _ = config.GetBool("erosd", "simulator")
	pythonPort, _ = config.GetString("python", "port")
	testMode, _ = config.GetBool("erosd", "testmode")
	allowsimulations, _ = config.GetBool("erosd", "allowsimulations")
	logPath, _ = config.GetString("erosd", "logpath")

	divisionCount, _ = config.GetInt64("ladderdivisions", "divisions")
	subdivisionCount, _ = config.GetInt64("ladderdivisions", "subdivisions")
	divisionPoints, _ = config.GetInt64("ladderdivisions", "subdivisionpoints")
	dn, err := config.GetString("ladderdivisions", "divisionnames")
	if err == nil {
		divisionNames = strings.Split(dn, ";")
	}

	matchmakingMatchTimeout, _ = config.GetInt64("ladder", "matchtimeout")
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

func listenAndServe(address string, admin bool) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("Failed to listen on", address, err)
	} else {
		if admin {
			log.Println("Listening admin on", address)
		} else {
			log.Println("Listening on", address)
		}
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error", err)
			continue
		}
		log.Println("Accepted", conn.RemoteAddr())

		if admin {
			go handleAdminConnection(conn)
		} else {
			go handleConnection(conn)
		}
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	err := loadConfig()
	if err != nil {
		log.Panicln("Error while loading config", err)
	}

	err = initDb()
	if err != nil {
		log.Fatalln("initDb", err)
	}

	if len(listenAddresses) == 0 {
		log.Fatalln("No listeners provided")
	}

	// start the broadcast routine
	go broadcastRunner()

	loadMaps()
	initDivisions()
	initClientCaches()
	initBattleNet()
	initMatchmaking()
	initChat()

	// start the matchmaker
	if simulator {
		// go matchmaker.run()
		doSimulations(25)
		return
	}

	// Set up listeners
	for _, listen := range listenAddresses {
		go listenAndServe(listen, false)
	}

	for _, listen := range adminListenAddresses {
		go listenAndServe(listen, true)
	}

	log.Println("Initialization complete")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP)

	for {
		sig := <-c
		switch sig {
		case os.Interrupt, os.Kill:
			SendBroadcastAlert(1, "")
			//os.Exit(0)
		case syscall.SIGHUP:
			break
		}
	}

}
