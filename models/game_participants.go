package models

/*
 *  Imports
 */

import (
	"database/sql"
)

/*
 *  Data types
 */

type GameParticipants struct{}

/*
 *  game_participants table functions
 */

func (*GameParticipants) GetCurrentGames(username string) ([]Game, error) {
	// Get a list of the non-finished games that the user is currently in
	rows, err := db.Query(`
		SELECT games.id, games.status
		FROM game_participants
			JOIN games ON game_participants.game_id = games.id
		WHERE game_participants.user_id = (SELECT id FROM users WHERE username = ?) AND games.status != 'finished'
		ORDER BY games.id
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the games
	var gameList []Game
	for rows.Next() {
		var game Game
		err := rows.Scan(&game.ID, &game.Status)
		if err != nil {
			return nil, err
		}

		// Append this game to the slice
		gameList = append(gameList, game)
	}

	return gameList, nil
}

func (*GameParticipants) GetNotStartedGames(userID int) ([]int, error) {
	// Get a list of the non-started games that the user is currently in
	rows, err := db.Query(`
		SELECT games.id
		FROM game_participants
			JOIN games ON game_participants.game_id = games.id
		WHERE game_participants.user_id = ? AND games.status == 'open'
		ORDER BY games.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the games
	var gameIDs []int
	for rows.Next() {
		var gameID int
		err := rows.Scan(&gameID)
		if err != nil {
			return nil, err
		}

		// Append this game to the slice
		gameIDs = append(gameIDs, gameID)
	}

	return gameIDs, nil
}

func (*GameParticipants) GetPlayerList(gameID int) ([]string, error) {
	// Get only the names of the people in this game
	rows, err := db.Query(`
		SELECT users.username
		FROM game_participants
			JOIN users ON users.id = game_participants.user_id
		WHERE game_participants.game_id = ?
		ORDER BY game_participants.id
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playerList []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		playerList = append(playerList, name)
	}

	return playerList, nil
}

func (*GameParticipants) CheckInGame(userID int, gameID int) (bool, error) {
	// Check to see if the user is in this game
	var id int
	err := db.QueryRow(`
		SELECT id
		FROM game_participants
		WHERE user_id = ? AND game_id = ?
	`, userID, gameID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (*GameParticipants) Insert(userID int, gameID int) error {
	// Add the user to the participants list for that game
	stmt, err := db.Prepare("INSERT INTO game_participants (user_id, game_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userID, gameID)
	if err != nil {
		return err
	}

	return nil
}

func (*GameParticipants) Delete(username string, gameID int) error {
	// Remove the user from the participants list for the respective game
	if stmt, err := db.Prepare(`
		DELETE FROM game_participants
		WHERE user_id = (SELECT id FROM users WHERE username = ?) AND game_id = ?
	`); err != nil {
		return err
	} else {
		_, err := stmt.Exec(username, gameID)
		if err != nil {
			return err
		}
	}

	// Get only the names of the people in this game (this is the same as the GameParticipants.GetPlayerNames function)
	rows, err := db.Query(`
		SELECT users.username
		FROM game_participants
			JOIN users ON users.id = game_participants.user_id
		WHERE game_participants.game_id = ?
		ORDER BY game_participants.id
	`, gameID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var playerNames []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return err
		}
		playerNames = append(playerNames, name)
	}

	// Check to see if anyone is still in this game
	if len(playerNames) == 0 {
		// Automatically close the game
		if stmt, err := db.Prepare("DELETE FROM games WHERE id = ?"); err != nil {
			return err
		} else {
			_, err := stmt.Exec(gameID)
			if err != nil {
				return err
			}
		}
	} else {
		// Check to see if this user was the captain
		var userID int
		if err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID); err != nil {
			return err
		}
		var captain int
		if err := db.QueryRow("SELECT captain FROM games WHERE id = ?", gameID).Scan(&captain); err != nil {
			return err
		}
		if userID == captain {
			// Change the captain to someone else
			stmt, err := db.Prepare(`
				UPDATE games
				SET captain = (SELECT user_id from game_participants WHERE game_id = ? ORDER BY id LIMIT 1)
				WHERE id = ?
			`)
			if err != nil {
				return err
			}
			_, err = stmt.Exec(gameID, gameID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
