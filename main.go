package main // In Go, executable commands must always use package main

/*
 *  Imports
 */

import (
	"github.com/Zamiell/hanabi-server/models"

	"net/http" // For establishing an HTTP server
	"os"       // For logging and reading environment variables
	"strconv"  // For converting the port number
	"sync"     // For locking and unlocking the connection map
	// For dealing with timestamps
	// For rate-limiting login requests
	"github.com/gorilla/context"     // For cookie sessions (1/2)
	"github.com/gorilla/sessions"    // For cookie sessions (2/2)
	"github.com/joho/godotenv"       // For reading environment variables that contain secrets
	"github.com/op/go-logging"       // For logging
	"github.com/tdewolff/minify"     // For minification (1/3)
	"github.com/tdewolff/minify/css" // For minification (2/3)
	"github.com/tdewolff/minify/js"  // For minification (3/3)
	"github.com/trevex/golem"        // The Golem WebSocket framework
)

/*
 *  Constants
 */

const (
	sessionName = "hanabi.sid"
	domain      = "hanabi.ddns.net"
	auth0Domain = "isaacserver.auth0.com"
	useSSL      = false
	sslCertFile = "/etc/letsencrypt/live/hanabi.ddns.net/fullchain.pem"
	sslKeyFile  = "/etc/letsencrypt/live/hanabi.ddns.net/privkey.pem"
)

/*
 *  Global variables
 */

var (
	projectPath   = os.Getenv("GOPATH") + "/src/github.com/Zamiell/hanabi-server"
	log           = logging.MustGetLogger("hanabi")
	sessionStore  *sessions.CookieStore
	roomManager   = golem.NewRoomManager()
	pmManager     = golem.NewRoomManager()
	commandMutex  = &sync.Mutex{} // Used to prevent race conditions
	connectionMap = make(map[string]*ExtendedConnection)
	chatRoomMap   = make(map[string][]User)
	gameMap       = make(map[int]*GameState)
	db            *models.Models
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
	err := godotenv.Load(projectPath + "/.env")
	if err != nil {
		log.Fatal("Failed to load .env file:", err)
	}

	// Create a session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	sessionStore = sessions.NewCookieStore([]byte(sessionSecret))
	maxAge := 60 * 60 * 24 * 30 // 1 month
	if useSSL == true {
		sessionStore.Options = &sessions.Options{
			Domain:   domain,
			Path:     "/",
			MaxAge:   maxAge,
			Secure:   true, // Only send the cookie over HTTPS: https://www.owasp.org/index.php/Testing_for_cookies_attributes_(OTG-SESS-002)
			HttpOnly: true, // Mitigate XSS attacks: https://www.owasp.org/index.php/HttpOnly
		}
	} else {
		sessionStore.Options = &sessions.Options{
			Domain:   domain,
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true, // Mitigate XSS attacks: https://www.owasp.org/index.php/HttpOnly
		}
	}

	// Initialize the database model
	if db, err = models.GetModels(projectPath + "/database.sqlite"); err != nil {
		log.Fatal("Failed to open the database:", err)
	}

	// Clean up any non-started games before we start
	if leftoverGames, err := db.Games.Cleanup(); err != nil {
		log.Fatal("Failed to cleanup the leftover games:", err)
	} else {
		for _, gameID := range leftoverGames {
			log.Info("Deleted game", gameID, "during starting cleanup.")
		}
	}

	// Initialize the achievements
	//achievementsInit()

	// Create a WebSocket router using the Golem framework
	router := golem.NewRouter()
	router.SetConnectionExtension(NewExtendedConnection)
	router.OnHandshake(validateSession)
	router.OnConnect(connOpen)
	router.OnClose(connClose)

	/*
	 *  The websocket commands
	 */

	// Chat commands
	router.On("roomJoin", roomJoin)
	router.On("roomLeave", roomLeave)
	router.On("roomMessage", roomMessage)
	router.On("privateMessage", privateMessage)

	// Game commands
	router.On("gameCreate", gameCreate)
	router.On("gameJoin", gameJoin)
	router.On("gameLeave", gameLeave)
	router.On("gameStart", gameStart)

	// Action commands
	router.On("actionPlay", actionPlay)
	router.On("actionClue", actionClue)
	router.On("actionDiscard", actionDiscard)

	// Profile commands
	/*router.On("profileGet", profileGet)
	router.On("profileSetUsername", profileSetUsername)*/

	// Admin commands
	/*router.On("adminBan", adminBan)
	router.On("adminUnban", adminUnban)
	router.On("adminBanIP", adminBanIP)
	router.On("adminUnbanIP", adminUnbanIP)
	router.On("adminSquelch", adminSquelch)
	router.On("adminUnsquelch", adminUnsquelch)
	router.On("adminPromote", adminPromote)
	router.On("adminDemote", adminDemote)*/

	/*
	 *  HTTP stuff
	 */

	// Minify CSS and JS
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	for _, fileName := range []string{"main"} {
		inputFile, _ := os.Open("public/css/" + fileName + ".css")
		outputFile, _ := os.Create("public/css/" + fileName + ".min.css")
		if err := m.Minify("text/css", outputFile, inputFile); err != nil {
			log.Error("Failed to minify \""+fileName+".css\":", err)
		}
	}
	m.AddFunc("text/javascript", js.Minify)
	for _, fileName := range []string{"login", "main"} {
		inputFile, _ := os.Open("public/js/" + fileName + ".js")
		outputFile, _ := os.Create("public/js/" + fileName + ".min.js")
		if err := m.Minify("text/javascript", outputFile, inputFile); err != nil {
			log.Error("Failed to minify \""+fileName+".js\":", err)
		}
	}

	// Assign functions to URIs
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public")))) // Serve static files
	http.HandleFunc("/", httpHandler)                                                          // Anything that is not a static file will match this
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/ws", router.Handler()) // The golem router handles websockets

	/*
	 *  Start the server
	 */

	// Figure out the port that we are using for the HTTP server
	var port int
	if useSSL == true {
		port = 443
	} else {
		port = 80
	}

	// Welcome message
	log.Info("Starting hanabi-server on port " + strconv.Itoa(port) + ".")

	// Listen and serve
	if useSSL == true {
		if err := http.ListenAndServeTLS(
			":"+strconv.Itoa(port), // Nothing before the colon implies 0.0.0.0
			sslCertFile,
			sslKeyFile,
			context.ClearHandler(http.DefaultServeMux), // We wrap with context.ClearHandler or else we will leak memory: http://www.gorillatoolkit.org/pkg/sessions
		); err != nil {
			log.Fatal("ListenAndServeTLS failed:", err)
		}
	} else {
		// Listen and serve (HTTP)
		if err := http.ListenAndServe(
			":"+strconv.Itoa(port),                     // Nothing before the colon implies 0.0.0.0
			context.ClearHandler(http.DefaultServeMux), // We wrap with context.ClearHandler or else we will leak memory: http://www.gorillatoolkit.org/pkg/sessions
		); err != nil {
			log.Fatal("ListenAndServe failed:", err)
		}
	}
}
