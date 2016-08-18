package main

/*
 *  Imports
 */

import (
	"github.com/Zamiell/hanabi-server/models"
	"github.com/trevex/golem"
	"time"
)

/*
 *  Golem data types
 */

// We must extend the default Golem connection so that it hold information about the user
type ExtendedConnection struct {
	Connection         *golem.Connection
	UserID             int
	Username           string
	Admin              int
	Squelched          int
	RateLimitAllowance float64
	RateLimitLastCheck time.Time
}

// Recieved in all commands
type IncomingCommandMessage struct {
	Room    string `json:"room"`
	Msg     string `json:"msg"`
	Name    string `json:"name"`
	Ruleset string `json:"ruleset"`
	ID      int    `json:"id"`
	Slot    int    `json:"slot"`
	Type    string `json:"type"`
	IP      string `json:"ip"`
}

// Sent in an "error" command (in the "connError" function)
type ErrorMessage struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// Sent upon connection to let the client know what their username is
type UsernameMessage struct {
	Name string `json:"name"`
}

/*
 *  Chat room data types
 */

// Sent in the "roomList" command to the person that is joining the room (in the "roomJoinSub" function)
type RoomListMessage struct {
	Room  string `json:"room"`
	Users []User `json:"users"`
}
type User struct {
	Name      string `json:"name"`
	Admin     int    `json:"admin"`
	Squelched int    `json:"squelched"`
}

// Sent in the "roomHistory" command to the person that is joining the room (in the "roomJoinSub" function)
type RoomHistoryMessage struct {
	Room    string               `json:"room"`
	History []models.RoomHistory `json:"history"`
}

// Sent in the "roomJoined" command to everyone who is already in the room (in the "roomJoinSub" function)
type RoomJoinedMessage struct {
	Room string `json:"room"`
	User User   `json:"user"`
}

// Sent in the "roomLeft" command (in the "roomLeaveSub" function)
type RoomLeftMessage struct {
	Room string `json:"room"`
	Name string `json:"name"`
}

// Sent in the "roomMessage" command (in the "roomMessage" function)
type RoomMessageMessage struct {
	Room string `json:"room"`
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

// Sent in the "privateMessage" command (in the "privateMessage" function)
type PrivateMessageMessage struct {
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

// Sent in the "roomSetName" command (in the "profileSetUsername" function)
type RoomSetNameMessage struct {
	Room    string `json:"room"`
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

// Sent in the "roomSetSquelched" command (in the "adminSquelch" and "adminUnsquelch" functions)
type RoomSetSquelchedMessage struct {
	Room      string `json:"room"`
	Name      string `json:"name"`
	Squelched int    `json:"squelched"`
}

// Sent in the "roomSetAdmin" command (in the "adminPromote" and "adminDemote" functions)
type RoomSetAdminMessage struct {
	Room  string `json:"room"`
	Name  string `json:"name"`
	Admin int    `json:"admin"`
}

/*
 *  Game data types
 */

// Sent in the "gameJoined" command (in the "gameJoin" function)
// Sent in the "gameLeft" command (in the "gameLeave" and "adminBan" functions)
type GameMessage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Sent in the "gameSetStatus" command (in the "gameCheckStart" functions)
type GameSetStatusMessage struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// Sent at the beginning of the game (or if joining a game late)
type HandListMessage struct {
	Name string `json:"name"`
	Hand Hand   `json:"hand"`
}

// Sent to tell the client that they got a new achievement
/*type AchievementMessage struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}*/

/*
 *  Action data types
 */

type ActionPlayDiscardMessage struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Slot     int      `json:"slot"`
	PlayCard PlayCard `json:"card"`
}

type ActionDrawMessage struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	PlayCard PlayCard `json:"card"`
}

/*
 *  Profile data types
 */

// Sent in the "profile" command (in the "getProfile" function)
/*type Profile struct {
	// TODO
}*/

/*
 *  Card data types
 */

// Used to keep track of the state of the game server-side
type GameState struct {
	Ruleset     string
	Turn        int
	Deck        []Card
	Hands       []Hand
	PlayPile    map[string]PlayCard // blue, green, yellow, red, purple, black/rainbow
	DiscardPile []PlayCard
	Clues       int
	Strikes     int
	ClueHistory []Clue
}
type Card struct {
	Color  string `json:"color"`
	Number int    `json:"number"`
}
type PlayCard struct {
	Card  Card   `json:"card"`
	Clues []Clue `json:"clues"`
}
type Hand struct {
	Name      string     `json:"name"`
	PlayerNum int        `json:"playerNum"`
	Cards     []PlayCard `json:"cards"` // max size of 5
}
type Clue struct {
	Turn          int    `json:"turn"`
	From          string `json:"from"`
	To            string `json:"to"`
	Type          string `json:"type"`
	AffectedSlots []int  `json:"affectedSlots"`
}

// Send in the "gameState" command (in the "connOpen" function)
type PlayerGameState struct {
	Ruleset     string              `json:"ruleset"`
	Turn        int                 `json:"turn"`
	DeckLeft    int                 `json:"deckLeft"`
	Hands       []Hand              `json:"hands"`
	PlayPile    map[string]PlayCard `json:"playPile"`
	DiscardPile []PlayCard          `json:"discardPile"`
	Clues       int                 `json:"clues"`
	Strikes     int                 `json:"strikes"`
	ClueHistory []Clue              `json:"clueHistory"`
}
