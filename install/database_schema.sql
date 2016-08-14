/*
 *  sqlite3 database.sqlite < install/database_schema.sql
 */

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id                INTEGER  PRIMARY KEY  AUTOINCREMENT,
    auth0_id          TEXT     NOT NULL,
    username          TEXT     NOT NULL,
    datetime_created  INTEGER  DEFAULT (strftime('%s', 'now')),
    last_login        INTEGER  DEFAULT (strftime('%s', 'now')),
    last_ip           TEXT     NOT NULL,
    admin             INTEGER  DEFAULT 0
);
CREATE UNIQUE INDEX users_index_auth0_id ON users (auth0_id);
CREATE UNIQUE INDEX users_index_username ON users (username COLLATE NOCASE);

DROP TABLE IF EXISTS games;
CREATE TABLE games (
    id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
    name                  TEXT                  DEFAULT "-",
    status                TEXT                  DEFAULT "open", /* in progress, finished */
    ruleset               TEXT                  DEFAULT "normal", /* black, rainbow */
    seed                  INTEGER               DEFAULT 0,
    datetime_created      INTEGER               DEFAULT (strftime('%s', 'now')),
    datetime_started      INTEGER               DEFAULT 0,
    datetime_finished     INTEGER               DEFAULT 0,
    captain               INTEGER               NOT NULL,
    FOREIGN KEY(captain)  REFERENCES users(id)
);
CREATE INDEX games_index_status ON games (status);
CREATE INDEX games_index_datetime_finished ON games (datetime_finished);

DROP TABLE IF EXISTS game_participants;
CREATE TABLE game_participants (
    id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
    user_id               INTEGER               NOT NULL,
    game_id               INTEGER               NOT NULL,
    datetime_joined       INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)  REFERENCES users(id),
    FOREIGN KEY(game_id)  REFERENCES games(id)
);
CREATE INDEX game_participants_index_user_id ON game_participants (user_id);
CREATE INDEX game_participants_index_game_id ON game_participants (game_id);
CREATE INDEX game_participants_index_datetime_joined ON game_participants (datetime_joined);

DROP TABLE IF EXISTS game_activity;
CREATE TABLE game_activity (
    id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
    user_id               INTEGER               NOT NULL,
    game_id               INTEGER               NOT NULL,
    datetime_action       INTEGER               DEFAULT (strftime('%s', 'now')),
    action_type           TEXT                  NOT NULL, /* play, clue, discard */
    play_card             TEXT                  DEFAULT "-", /* B1, G1, Y1, R1, P1, M1, L1 */
    clue_type             TEXT                  DEFAULT "-", /* B, G, Y, R, P, M, L, 1, 2, 3, 4, 5 */
    discard_slot          INTEGER               DEFAULT 0,
    FOREIGN KEY(user_id)  REFERENCES users(id),
    FOREIGN KEY(game_id)  REFERENCES games(id)
);
CREATE INDEX game_activity_index_user_id ON game_activity (user_id);
CREATE INDEX game_activity_index_game_id ON game_activity (game_id);
CREATE INDEX game_activity_index_datetime_action ON game_activity (datetime_action);

DROP TABLE IF EXISTS banned_users;
CREATE TABLE banned_users (
    id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
    user_id                         INTEGER               NOT NULL,
    admin_responsible               INTEGER               NOT NULL,
    datetime                        INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)            REFERENCES users(id),
    FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);
CREATE UNIQUE INDEX banned_users_index_user_id ON banned_users (user_id);

DROP TABLE IF EXISTS banned_ips;
CREATE TABLE banned_ips (
    id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
    ip                              TEXT                  NOT NULL,
    admin_responsible               INTEGER               NOT NULL,
    datetime                        INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);
CREATE UNIQUE INDEX banned_ips_index_ip ON banned_ips (ip);

DROP TABLE IF EXISTS squelched_users;
CREATE TABLE squelched_users (
    id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
    user_id                         INTEGER               NOT NULL,
    admin_responsible               INTEGER               NOT NULL,
    datetime                        INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)            REFERENCES users(id),
    FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);
CREATE UNIQUE INDEX squelched_users_index_user_id ON squelched_users (user_id);

DROP TABLE IF EXISTS chat_log;
CREATE TABLE chat_log (
    id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
    room                  TEXT                  NOT NULL,
    user_id               INTEGER               NOT NULL,
    message               TEXT                  NOT NULL,
    datetime              INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)  REFERENCES users(id)
);
CREATE INDEX chat_log_index_room ON chat_log (room);
CREATE INDEX chat_log_index_user_id ON chat_log (user_id);
CREATE INDEX chat_log_index_datetime ON chat_log (datetime);

DROP TABLE IF EXISTS chat_log_pm;
CREATE TABLE chat_log_pm (
    id                         INTEGER               PRIMARY KEY  AUTOINCREMENT,
    recipient_id               INTEGER               NOT NULL,
    user_id                    INTEGER               NOT NULL,
    message                    TEXT                  NOT NULL,
    datetime                   INTEGER               DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)       REFERENCES users(id),
    FOREIGN KEY(recipient_id)  REFERENCES users(id)
);
CREATE INDEX chat_log_pm_index_recipient_id ON chat_log_pm (recipient_id);
CREATE INDEX chat_log_pm_index_user_id ON chat_log_pm (user_id);
CREATE INDEX chat_log_pm_index_datetime ON chat_log_pm (datetime);

DROP TABLE IF EXISTS achievements;
CREATE TABLE achievements (
    id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
    name                  TEXT                  NOT NULL,
    description           TEXT                  NOT NULL
);
CREATE UNIQUE INDEX achievements_index_name ON achievements (name COLLATE NOCASE);

DROP TABLE IF EXISTS user_achievements;
CREATE TABLE user_achievements (
    id                           INTEGER                     PRIMARY KEY  AUTOINCREMENT,
    user_id                      INTEGER                     NOT NULL,
    achievement_id               INTEGER                     NOT NULL,
    datetime                     INTEGER                     DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY(user_id)         REFERENCES users(id),
    FOREIGN KEY(achievement_id)  REFERENCES achievement(id),
    UNIQUE(user_id, achievement_id)
);
CREATE INDEX user_achievements_index_user_id ON user_achievements (user_id);
CREATE INDEX user_achievements_index_achievement_id ON user_achievements (achievement_id);
