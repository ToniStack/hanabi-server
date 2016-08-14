package models

/*
 *  Imports
 */

import (
	"database/sql"                  // For connecting to the database (1/2)
	_ "github.com/mattn/go-sqlite3" // For connecting to the database (2/2)
)

/*
 *  Data types
 */

type Models struct {
	// Database tables
	//Achievements
	BannedIPs
	BannedUsers
	ChatLogPM
	ChatLog
	GameParticipants
	Games
	SquelchedUsers
	//UserAchievements
	Users
}

// Sent in the "roomHistory" command (in the "roomJoinSub" function)
type RoomHistory struct {
	Name     string `json:"name"`
	Msg      string `json:"msg"`
	Datetime int    `json:"datetime"`
}

// Sent in the "gameList" command (in the "connOpen" function)
// Sent in the "gameCreated" command (in the "gameCreate" function)
type Game struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Ruleset         string   `json:"ruleset"`
	DatetimeCreated int      `json:"datetime_created"`
	DatetimeStarted int      `json:"datetime_started"`
	Captain         string   `json:"captain"` // This is an integer in the database but we convert it to their name during the SELECT
	Players         []string `json:"players"`
}

/*
 *  Global variables
 */

var (
	db *sql.DB
)

/*
 *  Initialization function
 */

func GetModels(dbFile string) (*Models, error) {
	// Initialize the database
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints (which are disabled by default in SQLite3)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Create the model
	return &Models{}, nil
}
