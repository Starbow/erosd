package main

import (
	"database/sql"
	"errors"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbMap              *gorp.DbMap
	dbType             string
	dbConnectionString string
	ErrDbInsert        error = errors.New("An error occured while writing to the database.")
)

func initDb() (err error) {
	// connect to db
	db, err := sql.Open(dbType, dbConnectionString)
	if err != nil {
		return
	}
	// construct a gorp DbMap
	dbMap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add tables
	dbMap.AddTableWithName(Client{}, "clients").SetKeys(true, "Id")
	dbMap.AddTableWithName(BattleNetCharacter{}, "battle_net_characters").SetKeys(true, "Id")

	dbMap.AddTableWithName(MatchResult{}, "match_results").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchResultPlayer{}, "match_result_players").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchResultSource{}, "match_result_sources").SetKeys(true, "Id")

	dbMap.AddTableWithName(MatchmakerMatch{}, "matchmaker_matches").SetKeys(true, "Id")
	dbMap.AddTableWithName(MatchmakerMatchParticipant{}, "matchmaker_match_participants").SetKeys(true, "Id")

	dbMap.AddTableWithName(Map{}, "maps").SetKeys(true, "Id")
	dbMap.AddTableWithName(MapVeto{}, "map_vetoes").SetKeys(true, "Id")

	// create the tables.
	err = dbMap.CreateTablesIfNotExists()
	if err != nil {
		dbMap = nil
		return
	}

	return
}
