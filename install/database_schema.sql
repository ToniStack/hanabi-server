/* sqlite3 database.sqlite < install/database_schema.sql */

DROP TABLE IF EXISTS users;
CREATE TABLE users (
   id                INTEGER    PRIMARY KEY  AUTOINCREMENT,
   auth0_id          TEXT       NOT NULL,
   username          TEXT       NOT NULL,
   datetime_created  TIMESTAMP  DEFAULT (strftime('%s', 'now')),
   last_login        TIMESTAMP  DEFAULT (strftime('%s', 'now')),
   last_ip           TEXT       NOT NULL,
   admin             INTEGER    DEFAULT 0
);

DROP TABLE IF EXISTS games;
CREATE TABLE games (
   id                       INTEGER               PRIMARY KEY  AUTOINCREMENT,
   status                   TEXT                  DEFAULT "open", /* in progress, finished */
   ruleset                  TEXT                  DEFAULT "normal", /* black, rainbow */
   seed                     TEXT                  NOT NULL,
   datetime_created         TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   datetime_started         TIMESTAMP             DEFAULT 0,
   datetime_finished        TIMESTAMP             DEFAULT 0,
   created_by               INTEGER               NOT NULL,
   FOREIGN KEY(created_by)  REFERENCES users(id)
);
CREATE INDEX index_status ON races (status);

DROP TABLE IF EXISTS game_participants;
CREATE TABLE game_participants (
   id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
   user_id               INTEGER               NOT NULL,
   game_id               INTEGER               NOT NULL,
   datetime_joined       TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   FOREIGN KEY(user_id)  REFERENCES users(id),
   FOREIGN KEY(race_id)  REFERENCES races(id)
);

DROP TABLE IF EXISTS banned_ips;
CREATE TABLE banned_ips (
   id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
   ip                              TEXT                  NOT NULL,
   admin_responsible               INTEGER               NOT NULL,
   datetime                        TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);

DROP TABLE IF EXISTS banned_users;
CREATE TABLE banned_users (
   id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
   user_id                         INTEGER               NOT NULL,
   admin_responsible               INTEGER               NOT NULL,
   datetime                        TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   FOREIGN KEY(user_id)            REFERENCES users(id)
   FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);

DROP TABLE IF EXISTS squelched_users;
CREATE TABLE squelched_users (
   id                              INTEGER               PRIMARY KEY  AUTOINCREMENT,
   user_id                         INTEGER               NOT NULL,
   admin_responsible               INTEGER               NOT NULL,
   datetime                        TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   FOREIGN KEY(user_id)            REFERENCES users(id)
   FOREIGN KEY(admin_responsible)  REFERENCES users(id)
);

DROP TABLE IF EXISTS chat_log;
CREATE TABLE chat_log (
   id                    INTEGER               PRIMARY KEY  AUTOINCREMENT,
   user_id               INTEGER               NOT NULL,
   channel               TEXT                  NOT NULL,
   message               TEXT                  NOT NULL,
   datetime              TIMESTAMP             DEFAULT (strftime('%s', 'now')),
   FOREIGN KEY(user_id)  REFERENCES users(id)
);
