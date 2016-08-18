package models

/*
 *  Data types
 */

type GameActions struct{}

/*
 *  game_actions table functions
 */

func (*GameActions) InsertPlay(username string, gameID int, slot int) error {
	// Add the game action for this user
	stmt, err := db.Prepare(`
		INSERT INTO game_actions (user_id, game_id, action_type, slot)
		VALUES ((SELECT id FROM users WHERE username = ?), ?, 'play', ?)
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, gameID, slot)
	if err != nil {
		return err
	}

	return nil
}

func (*GameActions) InsertClue(username string, gameID int, clueType string) error {
	// Add the game action for this user
	stmt, err := db.Prepare(`
		INSERT INTO game_actions (user_id, game_id, action_type, clue_type)
		VALUES ((SELECT id FROM users WHERE username = ?), ?, 'clue', ?)
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, gameID, clueType)
	if err != nil {
		return err
	}

	return nil
}

func (*GameActions) InsertDiscard(username string, gameID int, slot int) error {
	// Add the game action for this user
	stmt, err := db.Prepare(`
		INSERT INTO game_actions (user_id, game_id, action_type, clue_type)
		VALUES ((SELECT id FROM users WHERE username = ?), ?, 'discard', ?)
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(username, gameID, slot)
	if err != nil {
		return err
	}

	return nil
}
