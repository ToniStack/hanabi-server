package main

/*
 *  Imports
 */

import (
	"github.com/Zamiell/hanabi-server/models"
	"strings"
)

/*
 *  WebSocket room/chat command functions
 */

func roomJoin(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "roomJoin"
	username := conn.Username
	room := data.Room

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command (for room \""+room+"\").")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate that the requested room is sane
	if room == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to join an empty room.")
		connError(conn, functionName, "That is not a valid room name.")
		return
	}

	// Validate that they are not trying to join a system room
	if strings.HasPrefix(room, "_") {
		commandMutex.Unlock()
		log.Warning("Access denied to system room.")
		connError(conn, functionName, "You are not allowed to manually join system rooms.")
		return
	}

	// Validate that they are not already in the room
	chatRoomMap.RLock()
	users, ok := chatRoomMap.m[room]
	chatRoomMap.RUnlock()
	if ok == true {
		// The room exists (at least 1 person is in it)
		userInRoom := false
		for _, user := range users {
			if user.Name == username {
				userInRoom = true
				break
			}
		}
		if userInRoom == true {
			commandMutex.Unlock()
			log.Warning("User \"" + username + "\" tried to join a room they were already in.")
			connError(conn, functionName, "You are already in that room.")
			return
		}
	}

	// Let them join the room
	roomJoinSub(conn, room)

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func roomLeave(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "roomLeave"
	username := conn.Username
	room := data.Room

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command (for room \""+room+"\").")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate that the requested room is sane
	if room == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to leave an empty room.")
		connError(conn, functionName, "That is not a valid room name.")
		return
	}

	// Validate that they are not trying to leave a system room
	if strings.HasPrefix(room, "_") {
		commandMutex.Unlock()
		log.Warning("Access denied to system room.")
		connError(conn, functionName, "You are not allowed to manually leave system rooms.")
		return
	}

	// Validate that the room exists
	chatRoomMap.RLock()
	users, ok := chatRoomMap.m[room]
	chatRoomMap.RUnlock()
	if ok == false {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to leave an invalid room.")
		connError(conn, functionName, "That is not a valid room name.")
		return
	}

	// Validate that they are actually in this room
	userInRoom := false
	for _, user := range users {
		if user.Name == username {
			userInRoom = true
			break
		}
	}
	if userInRoom == false {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to leave a room they were not in.")
		connError(conn, functionName, "You are not in that room.")
		return
	}

	// Let them leave the room
	roomLeaveSub(conn, room)

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func roomMessage(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "roomMessage"
	username := conn.Username
	room := data.Room
	msg := data.Msg

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Log the received command
	log.Debug("User \""+username+"\" sent a", functionName, "command (for room \""+room+"\").")

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate that the requested room is sane
	if room == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to message an empty room.")
		connError(conn, functionName, "That is not a valid room name.")
		return
	}

	// Strip leading and trailing whitespace from the message
	msg = strings.TrimSpace(msg)

	// Don't allow empty messages
	if msg == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to send an empty message.")
		connError(conn, functionName, "You cannot send an empty message.")
		return
	}

	// Validate that the user is not squelched
	if conn.Squelched == 1 {
		commandMutex.Unlock()
		connError(conn, functionName, "You have been squelched by an administrator, so you cannot chat with others.")
		return
	}

	// Validate that the room exists
	chatRoomMap.RLock()
	users, ok := chatRoomMap.m[room]
	chatRoomMap.RUnlock()
	if ok == false {
		commandMutex.Unlock()
		connError(conn, functionName, "That is not a valid room name.")
		return
	}

	// Validate that they are actually in this room
	userInRoom := false
	for _, user := range users {
		if user.Name == username {
			userInRoom = true
			break
		}
	}
	if userInRoom == false {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to message a room they were not in.")
		connError(conn, functionName, "You are not in that room.")
		return
	}

	// Add the new message to the database
	if err := db.ChatLog.Insert(room, username, msg); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Log the message
	log.Info("#" + room + " <" + username + "> " + msg)

	// Send the message
	roomManager.Emit(room, "roomMessage", &RoomMessageMessage{room, username, msg})

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

func privateMessage(conn *ExtendedConnection, data *IncomingCommandMessage) {
	// Local variables
	functionName := "privateMessage"
	username := conn.Username
	recipient := data.Name
	msg := data.Msg

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Rate limit all commands
	if commandRateLimit(conn) == true {
		return
	}

	// Validate that the requested person is sane
	if recipient == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to private message an empty string.")
		connError(conn, functionName, "That is not a valid person.")
		return
	}

	// Don't allow people to send PMs to themselves
	if recipient == username {
		commandMutex.Unlock()
		connError(conn, functionName, "You cannot send a private message to yourself.")
		return
	}

	// Strip leading and trailing whitespace from the message
	msg = strings.TrimSpace(msg)

	// Don't allow empty messages
	if msg == "" {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to send an empty message.")
		connError(conn, functionName, "You cannot send an empty message.")
		return
	}

	// Validate that the user is not squelched
	if conn.Squelched == 1 {
		commandMutex.Unlock()
		connError(conn, functionName, "You have been squelched by an administrator, so you cannot chat with others.")
		return
	}

	// Validate that the person is online
	connectionMap.RLock()
	_, ok := connectionMap.m[recipient]
	connectionMap.RUnlock()
	if ok == false {
		commandMutex.Unlock()
		log.Warning("User \"" + username + "\" tried to private message \"" + recipient + "\", who is offline.")
		connError(conn, functionName, "That user is not online.")
		return
	}

	// Add the new message to the database
	if err := db.ChatLogPM.Insert(recipient, username, msg); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		connError(conn, functionName, "Something went wrong. Please contact an administrator.")
		return
	}

	// Log the message
	log.Info("PM <" + username + "> <" + recipient + "> " + msg)

	// Send the message
	pmManager.Emit(recipient, "privateMessage", &PrivateMessageMessage{username, msg})

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

/*
 *  WebSocket room/chat subroutines
 */

func roomJoinSub(conn *ExtendedConnection, room string) {
	// Local variables
	username := conn.Username
	admin := conn.Admin
	squelched := conn.Squelched

	// Join the room
	roomManager.Join(room, conn.Connection)

	// Add the user to the chat room mapping
	chatRoomMap.Lock()
	userObject := User{username, admin, squelched}
	chatRoomMap.m[room] = append(chatRoomMap.m[room], userObject)
	users := chatRoomMap.m[room] // Save the list of users in the room for later
	chatRoomMap.Unlock()

	// Give the user the list of everyone in the chat room and tell everyone else that someone is joining
	connectionMap.RLock()
	for _, user := range users {
		userConnection, ok := connectionMap.m[user.Name]
		if ok == true { // All users in the chat room should be technically be online but there could be a race condition
			if user.Name == username {
				// Give the user the list of everyone in the chat room
				conn.Connection.Emit("roomList", &RoomListMessage{room, users})
			} else {
				// Send them a notification that someone else joined
				userConnection.Connection.Emit("roomJoined", &RoomJoinedMessage{room, userObject})
			}
		} else {
			log.Error("Failed to get the connection for user \"" + user.Name + "\" while connecting user \"" + username + "\" to room \"" + room + "\".")
			continue
		}
	}
	connectionMap.RUnlock()

	// Get the chat history for this channel
	var roomHistoryList []models.RoomHistory
	if strings.HasPrefix(room, "_game_") {
		// Get all of the history
		var err error
		roomHistoryList, err = db.ChatLog.Get(room, -1) // In SQLite, LIMIT -1 returns all results
		if err != nil {
			log.Error("Database error:", err)
			return
		}
	} else {
		// Get only the last 50 entries
		var err error
		roomHistoryList, err = db.ChatLog.Get(room, 50)
		if err != nil {
			log.Error("Database error:", err)
			return
		}
	}

	// Send the chat history
	conn.Connection.Emit("roomHistory", &RoomHistoryMessage{room, roomHistoryList})

	// Log the join
	log.Debug("User \"" + conn.Username + "\" joined room: #" + room)
}

func roomLeaveSub(conn *ExtendedConnection, room string) {
	// Local variables
	username := conn.Username

	// Leave the room
	roomManager.Leave(room, conn.Connection)

	// Get the index of the user in the chat room mapping for this room
	chatRoomMap.RLock()
	users, ok := chatRoomMap.m[room]
	chatRoomMap.RUnlock()
	if ok == false {
		log.Error("Failed to get the chat room map for room \"" + room + "\".")
		return
	}
	index := -1
	for i, user := range users {
		if user.Name == username {
			index = i
			break
		}
	}
	if index == -1 {
		log.Error("Failed to get the index for the current user in the chat room map for room \"" + room + "\".")
		return
	}

	// Remove the user from the chat room mapping
	chatRoomMap.Lock()
	chatRoomMap.m[room] = append(users[:index], users[index+1:]...)
	users = chatRoomMap.m[room]
	chatRoomMap.Unlock()

	// Since the amount of people in the chat room changed, send everyone an update
	connectionMap.RLock()
	for _, user := range users {
		userConnection, ok := connectionMap.m[user.Name] // This should always succeed, but there might be a race condition
		if ok == true {
			userConnection.Connection.Emit("roomLeft", &RoomLeftMessage{room, username})
		} else {
			log.Error("Failed to get the connection for user \"" + user.Name + "\" while disconnecting user \"" + username + "\" from room \"" + room + "\".")
			continue
		}
	}
	connectionMap.RUnlock()

	// Log the leave
	log.Debug("User \"" + conn.Username + "\" left room: #" + room)
}
