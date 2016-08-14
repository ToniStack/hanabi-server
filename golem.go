package main

/*
 *  Imports
 */

import (
	"github.com/trevex/golem" // The Golem WebSocket framework
	"net"                     // For splitting the IP address from the port
	"net/http"                // For establishing an HTTP server
	"time"                    // For rate limiting
)

/*
 *  Constants
 */

const (
	rateLimitRate = 30 // In commands sent
	rateLimitPer  = 60 // In seconds
)

/*
 *  Custom Golem connection constructor
 */

func NewExtendedConnection(conn *golem.Connection) *ExtendedConnection {
	return &ExtendedConnection{
		Connection: conn,
		UserID:     0, // These values will be set (again) during the connOpen function
		Username:   "",
		Admin:      0,
	}
}

/*
 *  Validate WebSocket connection
 */

func validateSession(w http.ResponseWriter, r *http.Request) bool {
	// Local variables
	functionName := "validateSession"
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Check to see if their IP is banned
	if userIsBanned, err := db.BannedIPs.Check(ip); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		return false
	} else if userIsBanned == true {
		commandMutex.Unlock()
		log.Info("IP \"" + ip + "\" tried to establish a WebSocket connection, but they are banned.")
		return false
	}

	// Get the session (this may be an empty session)
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Unable to get the session during the", functionName, "function:", err)
		return false
	}

	// If they have logged in, their cookie should have a "userID", "username", "admin", and "squelched" value
	if v, ok := session.Values["userID"]; ok == true && v.(int) > 0 {
		// Do nothing
	} else {
		commandMutex.Unlock()
		log.Debug("Unauthorized WebSocket handshake detected from:", ip, "(failed userID check)")
		return false
	}
	var username string
	if v, ok := session.Values["username"]; ok == true {
		username = v.(string)
	} else {
		commandMutex.Unlock()
		log.Debug("Unauthorized WebSocket handshake detected from:", ip, "(failed username check)")
		return false
	}
	if _, ok := session.Values["admin"]; ok == true {
		// Do nothing
	} else {
		commandMutex.Unlock()
		log.Debug("Unauthorized WebSocket handshake detected from:", ip, "(failed admin check)")
		return false
	}
	if _, ok := session.Values["squelched"]; ok == true {
		// Do nothing
	} else {
		commandMutex.Unlock()
		log.Debug("Unauthorized WebSocket handshake detected from:", ip, "(failed squelched check)")
		return false
	}

	// Check for sessions that belong to orphaned accounts
	if userExists, err := db.Users.Exists(username); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		return false
	} else if userExists == false {
		commandMutex.Unlock()
		log.Error("User \"" + username + "\" does not exist in the database; they are trying to establish a WebSocket connection with an orphaned account.")
		return false
	}

	// Check to see if this user is banned
	if userIsBanned, err := db.BannedUsers.Check(username); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		return false
	} else if userIsBanned == true {
		commandMutex.Unlock()
		log.Info("User \"" + username + "\" tried to log in, but they are banned.")
		return false
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()

	// If they got this far, they are a valid user
	return true
}

/*
 * Router connection functions
 */

func connOpen(conn *ExtendedConnection, r *http.Request) {
	// Local variables
	functionName := "connOpen"

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Get the session
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		// This should not fail, since we checked the session previously in the validateSession function
		commandMutex.Unlock()
		log.Error("Unable to get the session during the", functionName, "function:", err)
		return
	}

	// Get user information from the session
	var userID int
	if v, ok := session.Values["userID"]; ok == true && v.(int) > 0 {
		userID = v.(int)
	} else {
		commandMutex.Unlock()
		log.Error("Failed to retrieve \"userID\" from the session during the", functionName, "function.")
		return
	}
	var username string
	if v, ok := session.Values["username"]; ok == true {
		username = v.(string)
	} else {
		commandMutex.Unlock()
		log.Error("Failed to retrieve \"username\" from the session during the", functionName, "function.")
		return
	}
	var admin int
	if v, ok := session.Values["admin"]; ok == true {
		admin = v.(int)
	} else {
		commandMutex.Unlock()
		log.Error("Failed to retrieve \"admin\" from the session during the", functionName, "function.")
		return
	}
	var squelched int
	if v, ok := session.Values["squelched"]; ok == true {
		squelched = v.(int)
	} else {
		commandMutex.Unlock()
		log.Error("Failed to retrieve \"squelched\" from the cookie during the", functionName, "function.")
		return
	}

	// Store user information in the Golem connection so that we can use it in the Golem WebSocket functions later on
	conn.UserID = userID
	conn.Username = username
	conn.Admin = admin
	conn.Squelched = squelched
	conn.RateLimitAllowance = rateLimitRate
	conn.RateLimitLastCheck = time.Now()

	// Disconnect any existing connections with this username
	connectionMap.RLock()
	existingConnection, ok := connectionMap.m[username]
	connectionMap.RUnlock()
	if ok == true {
		log.Info("Closing existing connection for user \"" + username + "\".")
		connError(existingConnection, "logout", "You have logged on from somewhere else, so I'll disconnect you here.")
		existingConnection.Connection.Close()

		// Wait until the existing connection is terminated
		commandMutex.Unlock()
		for {
			connectionMap.RLock()
			_, ok := connectionMap.m[username]
			connectionMap.RUnlock()
			if ok == false {
				break
			}
		}
		commandMutex.Lock()
	}

	// Add the connection to a connection map so that we can keep track of all of the connections
	connectionMap.Lock()
	connectionMap.m[username] = conn
	connectionMap.Unlock()

	// Log the connection
	log.Info("User \""+username+"\" connected;", len(connectionMap.m), "user(s) now connected.")

	// Join the user to the global chat room
	roomJoinSub(conn, "global")

	// Join the user to the PMManager room corresponding to their username for private messages
	pmManager.Join(username, conn.Connection)

	// Get the current list of games
	gameList, err := db.Games.GetCurrentGames()
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Send it to the user
	conn.Connection.Emit("gameList", gameList)

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func connClose(conn *ExtendedConnection) {
	// Local variables
	userID := conn.UserID
	username := conn.Username

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Delete the connection from the connection map
	connectionMap.Lock()
	delete(connectionMap.m, username) // This will do nothing if the entry doesn't exist
	connectionMap.Unlock()

	// Make a list of all the chat rooms this person is in
	var chatRoomList []string
	chatRoomMap.RLock()
	for room, users := range chatRoomMap.m {
		for _, user := range users {
			if user.Name == username {
				chatRoomList = append(chatRoomList, room)
				break
			}
		}
	}
	chatRoomMap.RUnlock()

	// Leave all the chat rooms
	for _, room := range chatRoomList {
		roomLeaveSub(conn, room)
	}

	// Leave the chat room dedicated for private messages
	pmManager.LeaveAll(conn.Connection)

	// Check to see if this user is in any games that are not already in progress
	gameIDs, err := db.GameParticipants.GetNotStartedGames(userID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		return
	}

	// Iterate over the games that they are currently in
	for _, gameID := range gameIDs {
		// Remove this user from the participants list for that game
		if err := db.GameParticipants.Delete(username, gameID); err != nil {
			commandMutex.Unlock()
			log.Error("Database error:", err)
			return
		}

		// Send everyone a notification that the user left the game
		connectionMap.RLock()
		for _, conn := range connectionMap.m {
			conn.Connection.Emit("gameLeft", GameMessage{gameID, username})
		}
		connectionMap.RUnlock()
	}

	// Log the disconnection
	log.Info("User \""+username+"\" disconnected;", len(connectionMap.m), "user(s) now connected.")

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

/*
 *  WebSocket miscellaneous subroutines
 */

// Sent to the client if either their command was unsuccessful or something else went wrong
func connError(conn *ExtendedConnection, functionName string, msg string) {
	conn.Connection.Emit("error", &ErrorMessage{functionName, msg})
}

// Called at the beginning of every command handler
func commandRateLimit(conn *ExtendedConnection) bool {
	// Local variables
	username := conn.Username

	// Rate limit commands; algorithm from: http://stackoverflow.com/questions/667508/whats-a-good-rate-limiting-algorithm
	now := time.Now()
	timePassed := now.Sub(conn.RateLimitLastCheck).Seconds()
	conn.RateLimitLastCheck = now
	conn.RateLimitAllowance += timePassed * (rateLimitRate / rateLimitPer)
	if conn.RateLimitAllowance > rateLimitRate {
		conn.RateLimitAllowance = rateLimitRate
	}
	if conn.RateLimitAllowance < 1 {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" triggered rate-limiting; disconnecting them.")
		connError(conn, "logout", "You have been disconnected due to flooding.")
		conn.Connection.Close()
		return true
	} else {
		conn.RateLimitAllowance--
		return false
	}
}
