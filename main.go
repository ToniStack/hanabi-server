package main // In Go, executable commands must always use package main

/*
 *  Imports
 */

import (
	"net/http"                      // For establishing an HTTP server
	"os"                            // For logging and reading environment variables
	"sync"                          // For locking and unlocking the connection map
	"time"                          // For dealing with timestamps

	"github.com/op/go-logging"      // For logging
	"github.com/joho/godotenv"      // For reading environment variables that contain secrets
	"github.com/didip/tollbooth"    // For rate-limiting login requests
	"github.com/trevex/golem"       // The Golem WebSocket framework
	"github.com/gorilla/sessions"   // For cookie sessions (1/2)
	"github.com/gorilla/context"    // For cookie sessions (2/2)
	"database/sql"                  // For connecting to the database (1/2)
	_ "github.com/mattn/go-sqlite3" // For connecting to the database (2/2)
)

/*
 *  Constants
 */

const (
	port = "443"
	sessionName = "hanabi.sid"
	domainName = "isaacitemtracker.com"
)

/*
 *  Global variables
 */

var (
	log = logging.MustGetLogger("hanabi")
	sessionStore *sessions.CookieStore
	db *sql.DB
	roomManager = golem.NewRoomManager()
	connectionMap = struct {
		sync.RWMutex // Maps are not safe for concurrent use: https://blog.golang.org/go-maps-in-action
		m map[string]*ExtendedConnection
	}{m: make(map[string]*ExtendedConnection)}
)

/*
 *  Program entry point
 */

func main() {
	// Configure logging: http://godoc.org/github.com/op/go-logging#Formatter 
	loggingBackend := logging.NewLogBackend(os.Stdout, "", 0)
	logFormat := logging.MustStringFormatter( // https://golang.org/pkg/time/#Time.Format
		`%{time:Mon Jan 2 15:04:05 MST 2006} - %{level:.4s} - %{shortfile} - %{message}`,
	)
	loggingBackendFormatted := logging.NewBackendFormatter(loggingBackend, logFormat)
	logging.SetBackend(loggingBackendFormatted)

	// Load the .env file which contains environment variables with secret values
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env file:", err)
	}

	// Create a session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	sessionStore = sessions.NewCookieStore([]byte(sessionSecret))
	sessionStore.Options = &sessions.Options{
		Domain:   domainName,
		Path:     "/",
		MaxAge:   5, // 5 seconds
		Secure:   true, // Only send the cookie over HTTPS: https://www.owasp.org/index.php/Testing_for_cookies_attributes_(OTG-SESS-002)
		HttpOnly: true, // Mitigate XSS attacks: https://www.owasp.org/index.php/HttpOnly
	}

	// Initialize the database
	db, err = sql.Open("sqlite3", "database.sqlite")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create a WebSocket router using the Golem framework
	router := golem.NewRouter()
	router.SetConnectionExtension(NewExtendedConnection)
	router.OnHandshake(validateSession)
	router.OnConnect(connOpen)
	router.OnClose(connClose)

	// React on Golem room manager events
	//roomManager.On("create", roomCreated)
	//roomManager.On("remove", roomRemoved)

	/*
	 *  The websocket commands
	 */

	// Logout
	router.On("logout", logout)

	// Chat commands
	router.On("roomMessage", roomMessage)

	// Game commands (pre-starting)
	/*router.On("gameCreate", gameCreate)
	router.On("gameJoin", gameJoin)
	router.On("gameLeave", gameLeave)*/

	// Game commands (once started)
	/*router.On("gamePlayCard", gamePlayCard)
	router.On("gameClue", gameClue)
	router.On("gameDiscard", gameDiscard)*/

	/*
	 *  HTTP stuff
	 */

	// Assign functions to URIs
	http.HandleFunc("/ws", router.Handler())
	http.Handle("/login", tollbooth.LimitFuncHandler(tollbooth.NewLimiter(1, time.Second), loginHandler))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	// Welcome message
	log.Info("Starting hanabi-server on port " + port + ".")

	// Listen and serve (HTTPS)
	if err := http.ListenAndServeTLS(
		":" + port, // Nothing before the colon implies 0.0.0.0
		"/etc/letsencrypt/live/isaacitemtracker.com/fullchain.pem",
		"/etc/letsencrypt/live/isaacitemtracker.com/privkey.pem",
		context.ClearHandler(http.DefaultServeMux), // We wrap with context.ClearHandler or else we will leak memory: http://www.gorillatoolkit.org/pkg/sessions
	); err != nil {
		log.Fatal("ListenAndServeTLS failed:", err)
	}
}
