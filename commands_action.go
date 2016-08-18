package main

/*
 *  Imports
 */

import (
	"strconv"
)

/*
 *  WebSocket game action command functions
 */

func actionPlay(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "actionPlay"
	username := conn.Username
	gameID := data.ID
	slot := data.Slot

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

	// Validate that the game is in progress
	if gameValidateStatus(conn, data, "in progress", functionName) == false {
		return
	}

	// Validate that they are in the game
	if gameValidateIn(conn, data, functionName) == false {
		return
	}

	// Validate that it is their turn
	if actionValidateTurn(conn, data, functionName) == false {
		return
	}

	// Validate that the slot is not bogus
	if slot < 1 || slot > 5 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to discard a card from slot " + strconv.Itoa(slot) + ".")
		connError(conn, functionName, "Slot "+strconv.Itoa(slot)+" is an invalid slot.")
		return
	}

	// Validate that the player has a card in that slot
	var playerNum int
	for i, hand := range gameMap[gameID].Hands {
		if hand.Name == username {
			if len(hand.Cards) > slot {
				commandMutex.Unlock()
				log.Warning("Player \""+username+"\" tried to discard a card from slot "+strconv.Itoa(slot)+", but they only have", len(hand.Cards), "cards.")
				connError(conn, functionName, "You only have "+strconv.Itoa(len(hand.Cards))+" cards in your hand, so you can't discard from slot "+strconv.Itoa(slot)+".")
				return
			}

			playerNum = i
			break
		}
	}

	// Add the move to the database
	if err := db.GameActions.InsertPlay(username, gameID, slot); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Get the card at that slot and then remove it from their hand
	card := gameMap[gameID].Hands[playerNum].Cards[slot-1]
	gameMap[gameID].Hands[playerNum].Cards = gameMap[gameID].Hands[playerNum].Cards[slot:]

	// Check to see if the card is playable
	var command string
	if gameMap[gameID].PlayPile[card.Card.Color].Card.Number != card.Card.Number-1 {
		// Play the card
		gameMap[gameID].PlayPile[card.Card.Color] = card
		command = "actionPlay"
	} else {
		// Discard the card
		gameMap[gameID].DiscardPile = append(gameMap[gameID].DiscardPile, card)
		command = "actionMisplay"
	}

	// Tell everyone in the game the play/misplay
	for _, hand := range gameMap[gameID].Hands {
		conn, ok := connectionMap[hand.Name]
		if ok == true { // Not all players may be online during a game
			conn.Connection.Emit(command, &ActionPlayDiscardMessage{
				ID:       gameID,
				Name:     username,
				Slot:     slot,
				PlayCard: card,
			})
		}
	}

	// Get the top card from the deck
	newCard := PlayCard{
		Card:  gameMap[gameID].Deck[0],
		Clues: make([]Clue, 0),
	}
	gameMap[gameID].Deck = gameMap[gameID].Deck[1:]

	// Add it to the player's hand (but in the leftmost lost)
	// From: https://codingair.wordpress.com/2014/07/18/go-appendprepend-item-into-slice/
	gameMap[gameID].Hands[playerNum].Cards = append([]PlayCard{newCard}, gameMap[gameID].Hands[playerNum].Cards...)

	// Make a new unknown card
	unknownCard := PlayCard{
		Card: Card{
			Color:  "unknown",
			Number: 0,
		},
		Clues: make([]Clue, 0),
	}

	// Tell everyone in the game what they drew
	for _, hand := range gameMap[gameID].Hands {
		conn, ok := connectionMap[hand.Name]
		if ok == true { // Not all players may be online during a game
			if hand.Name == username {
				// Don't show the receiving player what they actually drew
				conn.Connection.Emit("actionDraw", &ActionDrawMessage{
					ID:       gameID,
					Name:     username,
					PlayCard: unknownCard,
				})
			} else {
				// Show everyone else the actual card
				conn.Connection.Emit("actionDraw", &ActionDrawMessage{
					ID:       gameID,
					Name:     username,
					PlayCard: newCard,
				})
			}
		}
	}

	// If a 5 was played, increment the clue count
	if card.Card.Number == 5 {
		gameMap[gameID].Clues++
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func actionClue(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "actionClue"
	username := conn.Username
	gameID := data.ID
	recipient := data.Name
	clueType := data.Type

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

	// Validate that the game is in progress
	if gameValidateStatus(conn, data, "in progress", functionName) == false {
		return
	}

	// Validate that they are in the game
	if gameValidateIn(conn, data, functionName) == false {
		return
	}

	// Validate that it is their turn
	if actionValidateTurn(conn, data, functionName) == false {
		return
	}

	// Validate that there is at least 1 clue left
	if gameMap[gameID].Clues == 0 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue at 0 clues.")
		connError(conn, functionName, "You cannot give a clue with 0 clues remaining.")
		return
	}

	// Validate that the clue recipient is in the game
	recipientNum := -1
	for i, hand := range gameMap[gameID].Hands {
		if hand.Name == recipient {
			recipientNum = i
			break
		}
	}
	if recipientNum == -1 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue to player \"" + recipient + "\", who is not in the game.")
		connError(conn, functionName, "That person is not in the game.")
		return
	}

	// Validate that the clue type is not bogus
	if clueType != "B" &&
		clueType != "G" &&
		clueType != "Y" &&
		clueType != "R" &&
		clueType != "P" &&
		clueType != "M" &&
		clueType != "K" &&
		clueType != "1" &&
		clueType != "2" &&
		clueType != "3" &&
		clueType != "4" &&
		clueType != "5" {

		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue of \"" + clueType + "\", which is bogus.")
		connError(conn, functionName, "That is not a valid clue type.")
		return
	}

	// Get the ruleset for this game
	ruleset, err := db.Games.GetRuleset(gameID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Validate the clue type is not bogus for specific rulesets
	if ruleset != "black" && clueType == "K" {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue of \"" + clueType + "\" in a non-black game.")
		connError(conn, functionName, "You can't give black clues in a non-black game.")
		return
	} else if ruleset != "rainbow" && clueType == "M" {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue of \"" + clueType + "\" in a non-rainbow game.")
		connError(conn, functionName, "You can't give rainbow clues in a non-rainbow game.")
		return
	}

	// Figure out the affected slots in the recipient's hand
	affectedSlots := make([]int, 0)
	for i, card := range gameMap[gameID].Hands[recipientNum].Cards {
		if (clueType == "B" && card.Card.Color == "blue") ||
			(clueType == "G" && card.Card.Color == "green") ||
			(clueType == "Y" && card.Card.Color == "yellow") ||
			(clueType == "R" && card.Card.Color == "red") ||
			(clueType == "P" && card.Card.Color == "purple") ||
			(clueType == "K" && card.Card.Color == "black") ||
			(clueType == "M" && card.Card.Color == "rainbow") ||
			(clueType == "1" && card.Card.Number == 1) ||
			(clueType == "2" && card.Card.Number == 2) ||
			(clueType == "3" && card.Card.Number == 3) ||
			(clueType == "4" && card.Card.Number == 4) ||
			(clueType == "5" && card.Card.Number == 5) {

			affectedSlots = append(affectedSlots, i+1)
		}
	}
	if len(affectedSlots) == 0 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to give a clue of \"" + clueType + "\" to \"" + recipient + "\", but they have no matching cards.")
		connError(conn, functionName, "You cannot give a clue unless it matches at least 1 card.")
		return
	}

	// Add the move to the database
	if err := db.GameActions.InsertClue(username, gameID, clueType); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Compile the clue
	clue := Clue{
		Turn:          gameMap[gameID].Turn,
		From:          username,
		To:            recipient,
		Type:          clueType,
		AffectedSlots: nil,
	}

	// Give the clue to the player
	gameMap[gameID].ClueHistory = append(gameMap[gameID].ClueHistory, clue)
	gameMap[gameID].Clues--

	// Tell everyone in the game that the clue happened
	for _, hand := range gameMap[gameID].Hands {
		conn, ok := connectionMap[hand.Name]
		if ok == true { // Not all players may be online during a game
			// The clue will be written to each player's clue history
			conn.Connection.Emit("actionClue", clue)
		}
	}

	// Decrement the clue count
	gameMap[gameID].Clues--

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func actionDiscard(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "actionDiscard"
	username := conn.Username
	gameID := data.ID
	slot := data.Slot + 1

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

	// Validate that the game is in progress
	if gameValidateStatus(conn, data, "in progress", functionName) == false {
		return
	}

	// Validate that they are in the game
	if gameValidateIn(conn, data, functionName) == false {
		return
	}

	// Validate that it is their turn
	if actionValidateTurn(conn, data, functionName) == false {
		return
	}

	// Validate that there is not 8 clues
	if gameMap[gameID].Clues == 8 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to discard at 8 clues.")
		connError(conn, functionName, "You cannot discard while at 8 clues.")
		return
	}

	// Validate that the slot is not bogus
	if slot < 1 || slot > 5 {
		commandMutex.Unlock()
		log.Warning("Player \"" + username + "\" tried to discard a card from slot " + strconv.Itoa(slot) + ".")
		connError(conn, functionName, "Slot "+strconv.Itoa(slot)+" is an invalid slot.")
		return
	}

	// Validate that the player has a card in that slot
	var playerNum int
	for i, hand := range gameMap[gameID].Hands {
		if hand.Name == username {
			if len(hand.Cards) > slot {
				commandMutex.Unlock()
				log.Warning("Player \""+username+"\" tried to discard a card from slot "+strconv.Itoa(slot)+", but they only have", len(hand.Cards), "cards.")
				connError(conn, functionName, "You only have "+strconv.Itoa(len(hand.Cards))+" cards in your hand, so you can't discard from slot "+strconv.Itoa(slot)+".")
				return
			}

			playerNum = i
			break
		}
	}

	// Add the move to the database
	if err := db.GameActions.InsertDiscard(username, gameID, slot); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Get the card at that slot and then remove it from their hand
	card := gameMap[gameID].Hands[playerNum].Cards[slot-1]
	gameMap[gameID].Hands[playerNum].Cards = gameMap[gameID].Hands[playerNum].Cards[slot:]

	// Discard the card
	gameMap[gameID].DiscardPile = append(gameMap[gameID].DiscardPile, card)

	// Increment the clue count
	gameMap[gameID].Clues++

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

/*
 *  Action subroutines
 */

func actionValidateTurn(conn *ExtendedConnection, data *IncomingCommandMessage, functionName string) bool {
	// Local variables
	username := conn.Username
	gameID := data.ID

	// Find out what player number this player is
	var playerNum int
	for _, hand := range gameMap[gameID].Hands {
		if hand.Name == username {
			playerNum = hand.PlayerNum
		}
	}

	// Validate that it is this player's turn
	if gameMap[gameID].Turn%len(gameMap[gameID].Hands) == 0 && playerNum == len(gameMap[gameID].Hands) {
		// It is the last player's turn; do nothing
	} else if gameMap[gameID].Turn%len(gameMap[gameID].Hands) == playerNum {
		// It is one of the other player's turns; do nothing
	} else {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" did a \"" + functionName + "\" when it was not their turn.")
		connError(conn, functionName, "You cannot perform a game action when it is not your turn.")
		return false
	}

	// It is the player's turn
	return true
}
