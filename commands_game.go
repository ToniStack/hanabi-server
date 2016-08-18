package main

/*
 *  Imports
 */

import (
	"github.com/Zamiell/hanabi-server/models"

	"math/rand"
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
		connAlert(conn, functionName, "There is already a non-finished game with that name.")
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
	for _, conn := range connectionMap {
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

	// Join the user to the channel for that game
	roomJoinSub(conn, "_game_"+strconv.Itoa(gameID))

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func gameJoin(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "gameJoin"
	userID := conn.UserID
	username := conn.Username
	gameID := data.ID

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command.")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate basic things about the game ID
	if gameValidate(conn, data, functionName) == false {
		return
	}

	// Validate that the game is open
	if gameValidateStatus(conn, data, "open", functionName) == false {
		return
	}

	// Validate that they are not in the game
	if gameValidateOut(conn, data, functionName) == false {
		return
	}

	// Add this user to the participants list for that game
	if err := db.GameParticipants.Insert(userID, gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Join the user to the channel for that game
	roomJoinSub(conn, "_game_"+strconv.Itoa(gameID))

	// Send everyone a notification that the user joined
	for _, conn := range connectionMap {
		conn.Connection.Emit("gameJoined", GameMessage{gameID, username})
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func gameLeave(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "gameLeave"
	username := conn.Username
	gameID := data.ID

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command.")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate basic things about the game ID
	if gameValidate(conn, data, functionName) == false {
		return
	}

	// Validate that the game is open
	if gameValidateStatus(conn, data, "open", functionName) == false {
		return
	}

	// Validate that they are in the game
	if gameValidateIn(conn, data, functionName) == false {
		return
	}

	// Remove this user from the participants list for that game
	if err := db.GameParticipants.Delete(username, gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Disconnect the user from the channel for that game
	roomLeaveSub(conn, "_game_"+strconv.Itoa(gameID))

	// Send everyone a notification that the user left the game
	for _, conn := range connectionMap {
		conn.Connection.Emit("gameLeft", GameMessage{gameID, username})
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func gameStart(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "gameStart"
	username := conn.Username
	userID := conn.UserID
	gameID := data.ID

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command.")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate basic things about the game ID
	if gameValidate(conn, data, functionName) == false {
		return
	}

	// Validate that the game is open
	if gameValidateStatus(conn, data, "open", functionName) == false {
		return
	}

	// Validate that they are in the game
	if gameValidateIn(conn, data, functionName) == false {
		return
	}

	// Validate that they are the captain of the game
	if isCaptain, err := db.Games.CheckCaptain(gameID, userID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	} else if isCaptain == false {
		commandMutex.Unlock()
		connError(conn, functionName, "Only the captain of a game can start it.")
		return
	}

	// Get the list of people in this game
	playerList, err := db.GameParticipants.GetPlayerList(gameID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Validate that there is between 2 and 5 people in the game
	if len(playerList) < 2 {
		commandMutex.Unlock()
		connAlert(conn, functionName, "You must have at least 2 people to start a game.")
		return
	} else if len(playerList) > 5 {
		commandMutex.Unlock()
		log.Error("Game #"+strconv.Itoa(gameID)+" tried to start but it somehow has", len(playerList), "people in it.")
		connError(conn, functionName, "You cannot start a game with more than 5 people.")
		return
	}

	// Change the status for this game to "in progress" and set "datetime_started" equal to now
	if err := db.Games.Start(gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Send everyone a notification that the game is now in progress
	for _, conn := range connectionMap {
		conn.Connection.Emit("gameSetStatus", &GameSetStatusMessage{gameID, "in progress"})
	}

	// Get the ruleset for this game
	ruleset, err := db.Games.GetRuleset(gameID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Create the deck
	colors := []string{"blue", "green", "yellow", "red", "purple"}
	if ruleset == "black" {
		colors = append(colors, "black")
	} else if ruleset == "rainbow" {
		colors = append(colors, "rainbow")
	}
	numbers := []int{1, 2, 3, 4, 5}
	var deck []Card
	for _, color := range colors {
		for _, number := range numbers {
			if number == 1 {
				// There are three 1's
				deck = append(deck, Card{color, number})
				deck = append(deck, Card{color, number})
				deck = append(deck, Card{color, number})

			} else if number == 2 || number == 3 || number == 4 {
				// There are two 2's, 3's, and 4's
				deck = append(deck, Card{color, number})
				deck = append(deck, Card{color, number})

			} else if number == 5 {
				// There is one 5
				deck = append(deck, Card{color, number})
			}
		}
	}

	// Get a seed for the game
	// TODO find seed that noone has played before
	seed := int64(rand.Intn(100000))
	log.Info("Seed chosen:", seed)

	// Add the seed to the database
	if err := db.Games.SetSeed(gameID, int(seed)); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Shuffle the deck based on this seed
	// From: http://stackoverflow.com/questions/12264789/shuffle-array-in-go
	r := rand.New(rand.NewSource(seed))
	for i := range deck {
		j := r.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	// Randomize the list of people
	for i := range playerList {
		j := rand.Intn(i + 1)
		playerList[i], playerList[j] = playerList[j], playerList[i]
	}

	// Initialize the game state
	gameMap[gameID] = &GameState{
		Ruleset:     ruleset,
		Turn:        1,
		PlayPile:    make(map[string]PlayCard),
		DiscardPile: make([]PlayCard, 0),
		Clues:       8,
		Strikes:     0,
		ClueHistory: make([]Clue, 0),
	}

	// Initialize the play pile part of the game state
	for _, color := range []string{"blue", "green", "yellow", "red", "purple"} {
		gameMap[gameID].PlayPile[color] = PlayCard{}
	}
	if ruleset == "black" {
		gameMap[gameID].PlayPile["black"] = PlayCard{}
	} else if ruleset == "rainbow" {
		gameMap[gameID].PlayPile["rainbow"] = PlayCard{}
	}

	// Initialize the player hands part of the game state
	for playerNum, player := range playerList {
		gameMap[gameID].Hands = append(gameMap[gameID].Hands, Hand{
			Name:      player,
			PlayerNum: playerNum + 1,
		})
	}

	// Figure out the hand size
	var handSize int
	if len(playerList) == 2 || len(playerList) == 3 {
		handSize = 5
	} else if len(playerList) == 4 || len(playerList) == 5 {
		handSize = 4
	} else {
		commandMutex.Unlock()
		log.Error("There is", len(playerList), "players in the game, which should never happen.")
		return
	}

	// Fill each player's hand and save the deck
	for i := 1; i <= handSize; i++ {
		for j := range playerList {
			// Get the top card from the deck
			card := PlayCard{
				Card:  deck[0],
				Clues: make([]Clue, 0),
			}
			deck = deck[1:]

			// Add it to the player's hand
			gameMap[gameID].Hands[j].Cards = append(gameMap[gameID].Hands[j].Cards, card)
		}
	}
	gameMap[gameID].Deck = deck

	// Send the players the game state
	for _, player := range playerList {
		conn, ok := connectionMap[player]
		if ok == true { // Not all players may be online during a game
			// Send them the game state
			gameSendState(conn, player, gameID)
		}
	}

	// Log the game starting
	log.Info("Game #" + strconv.Itoa(gameID) + " started.")

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()

	// Return for now and do more things in 2 hours
	go gameStart2(gameID)
}

func gameStart2(gameID int) {
	// Sleep 2 hours
	time.Sleep(2 * time.Hour)

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Send users a message that they took too long
	// TODO

	// Close the game
	// TODO

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

/*
 *  Game subroutines
 */

func gameValidate(conn *ExtendedConnection, data *IncomingCommandMessage, functionName string) bool {
	// Local variables
	username := conn.Username
	gameID := data.ID

	// Validate that the requested game is sane
	if gameID <= 0 {
		commandMutex.Unlock()
		log.Warning("User \""+username+"\" attempted to call", functionName, "with a bogus ID of "+strconv.Itoa(gameID)+".")
		connError(conn, functionName, "You must provide a valid game number.")
		return false
	}

	// Validate that the requested game exists
	if exists, err := db.Games.Exists(gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return false
	} else if exists == false {
		commandMutex.Unlock()
		log.Warning("User \""+username+"\" attempted to call", functionName, "on game #"+strconv.Itoa(gameID)+", but it doesn't exist.")
		connError(conn, functionName, "Game #"+strconv.Itoa(gameID)+" does not exist.")
		return false
	}

	// The user's request seems to be valid
	return true
}

func gameValidateIn(conn *ExtendedConnection, data *IncomingCommandMessage, functionName string) bool {
	// Local variables
	userID := conn.UserID
	username := conn.Username
	gameID := data.ID

	// Validate that they are in the game
	if userInGame, err := db.GameParticipants.CheckInGame(userID, gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return false
	} else if userInGame == false {
		commandMutex.Unlock()
		log.Warning("User \""+username+"\" attempted to call", functionName, "on game #"+strconv.Itoa(gameID)+", but they are not in that game.")
		connError(conn, functionName, "You are not in game #"+strconv.Itoa(gameID)+".")
		return false
	}

	// The user is in the game
	return true
}

func gameValidateOut(conn *ExtendedConnection, data *IncomingCommandMessage, functionName string) bool {
	// Local variables
	userID := conn.UserID
	username := conn.Username
	gameID := data.ID

	// Validate that they are not already in the game
	if userInGame, err := db.GameParticipants.CheckInGame(userID, gameID); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return false
	} else if userInGame == true {
		commandMutex.Unlock()
		log.Warning("User \""+username+"\" attempted to call", functionName, "on game #"+strconv.Itoa(gameID)+", but they are already in that game.")
		connError(conn, functionName, "You are already in game #"+strconv.Itoa(gameID)+".")
		return false
	}

	// The user is not in the game
	return true
}

func gameValidateStatus(conn *ExtendedConnection, data *IncomingCommandMessage, status string, functionName string) bool {
	// Local variables
	username := conn.Username
	gameID := data.ID

	// Validate that the game is set to the correct status
	if correctStatus, err := db.Games.CheckStatus(gameID, status); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return false
	} else if correctStatus == false {
		commandMutex.Unlock()
		log.Warning("User \""+username+"\" attempted to call", functionName, "on game #"+strconv.Itoa(gameID)+", but game is not set to status \""+status+"\".")
		connError(conn, functionName, "Game #"+strconv.Itoa(gameID)+" is not set to status \""+status+"\".")
		return false
	}

	// The game is the correct status
	return true
}

func gameSendState(conn *ExtendedConnection, username string, gameID int) {
	// Compile the state of the game thusfar
	gameState := PlayerGameState{
		Ruleset:     gameMap[gameID].Ruleset,
		Turn:        gameMap[gameID].Turn,
		DeckLeft:    len(gameMap[gameID].Deck),
		Hands:       make([]Hand, len(gameMap[gameID].Hands)),
		PlayPile:    gameMap[gameID].PlayPile,
		DiscardPile: gameMap[gameID].DiscardPile,
		Clues:       gameMap[gameID].Clues,
		Strikes:     gameMap[gameID].Strikes,
		ClueHistory: gameMap[gameID].ClueHistory,
	}

	// Copy the hands by value instead of by reference (because we need to modify it)
	copy(gameState.Hands, gameMap[gameID].Hands)
	for i := range gameState.Hands {
		// Since we need to modify a slice within a slice, we have to copy 2 levels deep
		gameState.Hands[i].Cards = make([]PlayCard, len(gameState.Hands[i].Cards))
		copy(gameState.Hands[i].Cards, gameMap[gameID].Hands[i].Cards)

		// Modify the hand so that we don't show a player their own cards
		if gameState.Hands[i].Name == username {
			for j := range gameState.Hands[i].Cards {
				gameState.Hands[i].Cards[j].Card.Color = "unknown"
				gameState.Hands[i].Cards[j].Card.Number = 0
			}
		}
	}

	// Send the game state to the player
	conn.Connection.Emit("gameState", gameState)
}
