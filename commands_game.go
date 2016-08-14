package main

/*
 *  Imports
 */

import (
	"github.com/Zamiell/hanabi-server/models"

	"strconv"
	"time"
)

/*
 *  WebSocket game command functions
 */

func gameCreate(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "gameCreate"
	userID := conn.UserID
	username := conn.Username
	name := data.Name
	ruleset := data.Ruleset

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command.")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate that the game name cannot be empty
	if name == "" {
		name = "-"
	}

	// Validate that the ruleset cannot be empty
	if ruleset == "" {
		ruleset = "normal"
	}

	// Validate the submitted ruleset
	if ruleset != "normal" && ruleset != "black" && ruleset != "rainbow" {
		commandMutex.Unlock()
		connError(conn, functionName, "That is not a valid ruleset.")
		return
	}

	// Check if this user is already in a game
	if gameList, err := db.GameParticipants.GetCurrentGames(username); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	} else if len(gameList) > 0 {
		commandMutex.Unlock()
		log.Info("New game request denied; \"" + username + "\" is already in a game.")
		connError(conn, functionName, "You can't create a new game if you are already in one.")
		return
	}

	// Check if there are any non-finished games with the same name
	if gameWithSameName, err := db.Games.CheckName(name); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	} else if gameWithSameName == true {
		commandMutex.Unlock()
		connError(conn, functionName, "There is already a non-finished game with that name.")
		return
	}

	// Create the game
	gameID, err := db.Games.Insert(name, ruleset, userID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Add this user to the participants list for that game
	if err := db.GameParticipants.Insert(userID, gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Send everyone a notification that a new game has been started
	connectionMap.RLock()
	for _, conn := range connectionMap.m {
		conn.Connection.Emit("gameCreated", &models.Game{
			ID:              gameID,
			Name:            name,
			Status:          "open",
			Ruleset:         ruleset,
			DatetimeCreated: int(time.Now().Unix()),
			Captain:         username,
			Players:         []string{username},
		})
	}
	connectionMap.RUnlock()

	// Join the user to the channel for that game
	roomJoinSub(conn, "_game_"+strconv.Itoa(gameID))

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}
