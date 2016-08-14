package models

/*
 *  Imports
 */

import (
	"strings"
)

/*
 *  Data types
 */

type Games struct{}

/*
 *  games table functions
 */

func (*Games) GetCurrentGames() ([]Game, error) {
	// Get the current games
	rows, err := db.Query(`
		SELECT id, name, status,
			ruleset, datetime_created, datetime_started,
			(SELECT username FROM users WHERE id = captain) as captain
		FROM games
		WHERE status != 'finished'
		ORDER BY datetime_created
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// We have to initialize this way to avoid sending a null on an empty array: https://danott.co/posts/json-marshalling-empty-slices-to-empty-arrays-in-go.html
	gameList := make([]Game, 0)
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&game.ID, &game.Name, &game.Status,
			&game.Ruleset, &game.DatetimeCreated, &game.DatetimeStarted,
			&game.Captain,
		)
		if err != nil {
			return nil, err
		}

		// Add it to the list
		gameList = append(gameList, game)
	}

	// Get the names of the people in this game
	rows, err = db.Query(`
		SELECT games.id, users.username
		FROM games
			JOIN game_participants ON game_participants.game_id = games.id
			JOIN users ON users.id = game_participants.user_id
		WHERE games.status != 'finished'
		ORDER BY game_participants.datetime_joined
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// For each name that we found, append it to the appropriate place in the gameList object
	for rows.Next() {
		var gameID int
		var username string
		err := rows.Scan(&gameID, &username)
		if err != nil {
			return nil, err
		}

		// Find the game in the gameList object
		for i, game := range gameList {
			if game.ID == gameID {
				gameList[i].Players = append(game.Players, username)
				break
			}
		}
	}

	return gameList, nil
}

func (*Games) CheckName(name string) (bool, error) {
	// Check to see if there are non-finished games with the same name
	rows, err := db.Query("SELECT name FROM games WHERE status != 'finished'")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var gameName string
		err := rows.Scan(&gameName)
		if err != nil {
			return false, err
		}

		if strings.ToLower(name) == strings.ToLower(gameName) && name != "-" {
			return true, nil
		}
	}

	return false, nil
}

func (*Games) Insert(name string, ruleset string, userID int) (int, error) {
	// Add the game to the database
	var gameID int
	if stmt, err := db.Prepare(`
		INSERT INTO games (name, ruleset, captain)
		VALUES (?, ?, ?)
	`); err != nil {
		return 0, err
	} else {
		result, err := stmt.Exec(name, ruleset, userID)
		if err != nil {
			return 0, err
		}
		gameID64, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		gameID = int(gameID64)
	}

	return gameID, nil
}

func (*Games) Cleanup() ([]int, error) {
	// Get the current games
	rows, err := db.Query("SELECT id FROM games WHERE status = 'open' ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the current games
	var leftoverGames []int
	for rows.Next() {
		var gameID int
		err := rows.Scan(&gameID)
		if err != nil {
			return nil, err
		}

		leftoverGames = append(leftoverGames, gameID)
	}

	// Delete all of the entries from the game_participants table (we don't want to use GameParticipants.Delete because we don't care about captains)
	for _, gameID := range leftoverGames {
		stmt, err := db.Prepare("DELETE FROM game_participants WHERE game_id = ?")
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(gameID)
		if err != nil {
			return nil, err
		}
	}

	// Delete the entries from the games table
	for _, gameID := range leftoverGames {
		stmt, err := db.Prepare("DELETE FROM games WHERE id = ?")
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(gameID)
		if err != nil {
			return nil, err
		}
	}

	return leftoverGames, nil
}