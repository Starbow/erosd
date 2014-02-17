package main

import (
	"database/sql"
	"errors"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var (
	dbMap              *gorp.DbMap
	dbType             string
	dbConnectionString string
	ErrDbInsert        error = errors.New("An error occured while writing to the database.")
)

// Website user in the database
type RealUser struct {
	Id        int64  `db:"id"`
	Username  string `db:"username"`
	AuthToken string `db:"authtoken"`
	Email     string `db:"email"`
}

// Attempt to log in. Should probably rate limit this.
func GetRealUser(username, authtoken string) *RealUser {
	if username == "" || authtoken == "" {
		return nil
	}
	var user RealUser
	err := dbMap.SelectOne(&user, "SELECT id, username, authtoken FROM user_user WHERE username=?", username)
	if testMode {
		// Test mode creates the user if it doesn't exist.
		if err != nil || user.Id == 0 {
			user.Username = username
			user.AuthToken = authtoken
			user.Email = username + "@starbowmod.com"
			err = dbMap.Insert(&user)
			if err != nil {
				log.Println(err)
			}
		}

	}

	if err != nil || user.Id == 0 {
		return nil
	}

	if user.AuthToken != authtoken {
		return nil
	}

	return &user
}

func initDb() (err error) {
	// connect to db
	db, err := sql.Open(dbType, dbConnectionString)
	if err != nil {
		return
	}
	// construct a gorp DbMap
	switch dbType {
	case "sqlite3":
		dbMap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	case "mysql":
		dbMap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	}

	// add tables
	dbMap.AddTableWithName(Client{}, "clients").SetKeys(false, "Id")
	dbMap.AddTableWithName(ClientRegionStats{}, "client_region_stats").SetKeys(true, "Id")
	dbMap.AddTableWithName(BattleNetCharacter{}, "battle_net_characters").SetKeys(true, "Id")

	dbMap.AddTableWithName(MatchResult{}, "match_results").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchResultPlayer{}, "match_result_players").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchResultSource{}, "match_result_sources").SetKeys(true, "Id")

	dbMap.AddTableWithName(MatchmakerMatch{}, "matchmaker_matches").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchmakerMatchParticipant{}, "matchmaker_match_participants").SetKeys(true, "Id")

	dbMap.AddTableWithName(Map{}, "maps").SetKeys(true, "Id")
	dbMap.AddTableWithName(MapVeto{}, "map_vetoes").SetKeys(true, "Id")

	if testMode {
		dbMap.AddTableWithName(RealUser{}, "user_user").SetKeys(true, "Id")
	}

	// create the tables.
	err = dbMap.CreateTablesIfNotExists()
	if err != nil {
		dbMap = nil
		return
	}

	return
}
