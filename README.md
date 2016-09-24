hanabi-server
=============

Description
-----------

This is a server intended to replace the old keldon.net Hanabi server. But this time, it will be open source!

The server is written in Go and uses WebSockets to communicate with the client. It leverages Auth0 for authentication and uses a SQLite database to keep track of the games.



Install
-------

* Install Go (you need to be able to run the `go` command).
* Install SQLite3 (you need to be able to run the `sqlite3` command).
* `go get github.com/Zamiell/hanabi-server`
* `cd $GOPATH/Zamiell/hanabi-server`
* `sqlite3 database.sqlite < install/database_schema.sql`
* Open up the `main.go` file and change the constants near the top of the file to your liking.
* Create a `.env` file in the current directory with the following contents:

```
SESSION_SECRET=some_long_random_string
AUTH0_CLIENT_ID=the_client_id_from_auth0
AUTH0_CLIENT_SECRET=the_client_secret_from_auth0
```



Run
---

* `cd $GOPATH/Zamiell/hanabi-server`
* `go run *.go`
