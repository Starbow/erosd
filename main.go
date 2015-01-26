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
	httpListenAddresses  []string
	httpsListenAddresses []string
	simulator            bool
	allowsimulations     bool
	matchmaker           *Matchmaker
	testMode             bool
	logPath              string
	replayPath           string
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
		config.AddOption("erosd", "httplisten", ":9090")
		config.AddOption("erosd", "werbroot", "web/")
		config.AddOption("erosd", "adminlisten", "127.0.0.1:12346")
		config.AddOption("erosd", "simulator", "false")
		config.AddOption("erosd", "allowsimulations", "false")
		config.AddOption("erosd", "testmode", "false")
		config.AddOption("erosd", "logpath", "logs")
		config.AddOption("erosd", "replaypath", "replays")

		// HTTPS
		config.AddOption("erosd", "httpslisten", ":9091")
		config.AddOption("erosd", "tlscertpath", "server.crt")
		config.AddOption("erosd", "tlskeypath", "server.key")

		config.AddSection("ladderdivisions")
		config.AddOption("ladderdivisions", "divisions", "4")
		config.AddOption("ladderdivisions", "divisionincrements", "5")
		config.AddOption("ladderdivisions", "divisionfirst", "20")
		config.AddOption("ladderdivisions", "divisionnames", "E;D;C;B;A")

		config.AddSection("ladder")
		config.AddOption("ladder", "longprocessresponsetime", "240")
		config.AddOption("ladder", "longprocessunlocktime", "60")
		config.AddOption("ladder", "startingpoints", "0")
		config.AddOption("ladder", "winpointsbase", "100")
		config.AddOption("ladder", "winpointsincrement", "25")
		config.AddOption("ladder", "losepointsbase", "50")
		config.AddOption("ladder", "losepointsincrement", "12.5")
		config.AddOption("ladder", "maxvetos", "3")
		config.AddOption("ladder", "ratingscale", "0.12")
		config.AddOption("ladder", "radiusmultiplier", "8.00")

		config.AddSection("database")
		config.AddOption("database", "type", "sqlite3")
		config.AddOption("database", "connection", "erosd.sqlite3")

		config.AddSection("chat")
		config.AddOption("chat", "fixedrooms", "Starbow;Practice Partner Search (Bronze-Silver);Practice Partner Search (Gold-Platinum);Practice Partner Search")
		config.AddOption("chat", "autojoin", "autojoin;Practice Partner Search (Bronze-Silver);Practice Partner Search (Gold-Platinum);Practice Partner Search")
		config.AddOption("chat", "maxuserchats", "5")
		config.AddOption("chat", "delay", "250")
		config.AddOption("chat", "maxthrottletime", "300")
		config.AddOption("chat", "maxmessagelength", "256")
		config.AddOption("chat", "maxmessagecache", "500")

		config.AddSection("python")
		config.AddOption("python", "port", ":54321")

		config.AddSection("oauth2")
		config.AddOption("oauth2", "clientid", "")
		config.AddOption("oauth2", "clientsecret", "")
		config.AddOption("oauth2", "codetimeoutminutes", "10")
		config.AddOption("oauth2", "redirecturi", "https://eros.starbowmod.com/login/battlenet")

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
	listen, err = config.GetString("erosd", "httplisten")
	if err == nil {
		httpListenAddresses = strings.Split(listen, ";")
	}
	listen, err = config.GetString("erosd", "httpslisten")
	if err == nil {
		httpsListenAddresses = strings.Split(listen, ";")
	}
	simulator, _ = config.GetBool("erosd", "simulator")
	pythonPort, _ = config.GetString("python", "port")
	testMode, _ = config.GetBool("erosd", "testmode")
	allowsimulations, _ = config.GetBool("erosd", "allowsimulations")
	logPath, _ = config.GetString("erosd", "logpath")
	replayPath, _ = config.GetString("erosd", "replaypath")
	webRoot, _ = config.GetString("erosd", "webroot")
	tlsCertPath, _ = config.GetString("erosd", "tlscertpath")
	tlsKeyPath, _ = config.GetString("erosd", "tlskeypath")

	if webRoot == "" {
		webRoot = "web/"
	}

	divisionCount, _ = config.GetInt64("ladderdivisions", "divisions")
	divisionIncrements, _ = config.GetFloat("ladderdivisions", "divisionincrements")
	divisionFirstRating, _ = config.GetFloat("ladderdivisions", "divisionfirst")
	dn, err := config.GetString("ladderdivisions", "divisionnames")
	if err == nil {
		divisionNames = strings.Split(dn, ";")
	}

	matchmakingMatchTimeout, _ = config.GetInt64("ladder", "matchtimeout")
	matchmakingLongProcessUnlockTime, _ = config.GetInt64("ladder", "longprocessunlocktime")
	matchmakingLongProcessResponseTime, _ = config.GetInt64("ladder", "longprocessresponsetime")

	matchmakingRadiusMultiplier, _ = config.GetFloat("ladder", "radiusmultiplier")
	matchmakingRatingScalePerSecond, _ = config.GetFloat("ladder", "ratingscale")

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
	aj, err := config.GetString("chat", "autojoin")
	if err == nil {
		autoJoinChatRooms = strings.Split(aj, ";")
	}
	maxChatRooms, _ = config.GetInt64("chat", "maxuserchats")
	chatMaxMessageLength, _ = config.GetInt64("chat", "maxmessagelength")
	mtt, _ := config.GetInt64("chat", "maxthrottletime")
	chatMaxThrottleTime = time.Duration(mtt) * time.Second
	delay, _ := config.GetInt64("chat", "delay")
	chatDelay = time.Duration(delay) * time.Millisecond

	maxMessageCache64, _ := config.GetInt64("chat", "maxmessagecache")
	maxMessageCache = int(maxMessageCache64)

	// OAuth
	oauthClientId, _ = config.GetString("oauth2", "clientid")
	oauthClientSecret, _ = config.GetString("oauth2", "clientsecret")
	oauthCodeTimeout, _ = config.GetInt64("oauth2", "codetimeoutminutes")
	oauthRedirectUri, _ = config.GetString("oauth2", "redirecturi")

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

	for _, listen := range httpListenAddresses {
		go listenAndServeHTTP(listen)
	}

	for _, listen := range httpsListenAddresses {
		go listenAndServeHTTPS(listen)
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
