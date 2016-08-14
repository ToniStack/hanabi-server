package main

/*
 *  Imports
 */

import (
	"html/template"
	"net"
	"net/http"
	"os"

	"encoding/json"       // For Auth0 authentication (1/3)
	"golang.org/x/oauth2" // For Auth0 authentication (2/3)
	"io/ioutil"           // For Auth0 authentication (3/3)
)

/*
 *  HTTP functions
 */

func httpHandler(w http.ResponseWriter, r *http.Request) {
	// Local variables
	functionName := "httpHandler"
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Check to see if their IP is banned
	if userIsBanned, err := db.BannedIPs.Check(ip); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if userIsBanned == true {
		commandMutex.Unlock()
		log.Info("IP \"" + ip + "\" tried to view the website, but they are banned.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get the session (this may be an empty session)
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Unable to get the session during the", functionName, "function:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check to see if they are already logged in
	var loggedIn bool
	if _, ok := session.Values["userID"]; ok == true {
		loggedIn = true
	} else {
		loggedIn = false
	}

	// Select either the login template or the main template
	lp := projectPath + "/views/layout.tmpl"
	var fp string
	if loggedIn == false {
		fp = projectPath + "/views/login.tmpl"
	} else {
		fp = projectPath + "/views/main.tmpl"
	}

	// Create the template
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Error("Failed to create the template:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Execute the template and send it to the user
	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Error("Failed to execute the template:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()
}

/*
 *  Login users using Auth0 access tokens
 */

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Local variables
	functionName := "loginHandler"
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Check to see if their IP is banned
	if userIsBanned, err := db.BannedIPs.Check(ip); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if userIsBanned == true {
		commandMutex.Unlock()
		log.Info("IP \"" + ip + "\" tried to log in, but they are banned.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get the session (this may be an empty session)
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Unable to get the session during the", functionName, "function:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check to see if they are already logged in (which should probably never happen since the cookie is so short)
	if _, ok := session.Values["userID"]; ok == true {
		commandMutex.Unlock()
		log.Warning("User from IP \"" + ip + "\" tried to get a session cookie, but they are already logged in.")
		http.Error(w, "You are already logged in. Please wait 10 seconds.", http.StatusUnauthorized)
		return
	}

	// Instantiate the OAuth2 package
	var redirectURL string
	if useSSL == true {
		redirectURL = "https://"
	} else {
		redirectURL = "http://"
	}
	redirectURL += domain + "/login"
	conf := &oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "name", "email", "nickname"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + auth0Domain + "/authorize",
			TokenURL: "https://" + auth0Domain + "/oauth/token",
		},
	}

	// Validate the code that the user sent from Auth0
	code := r.URL.Query().Get("code")
	if code == "" {
		commandMutex.Unlock()
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Exchanging the code for a token
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Failed to exchange the code of \""+code+"\" for a token:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Get information about the user
	client := conf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://" + auth0Domain + "/userinfo")
	if err != nil {
		commandMutex.Unlock()
		log.Error("Failed to login with Auth0 token \""+token.AccessToken+"\":", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Reading the body
	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		commandMutex.Unlock()
		log.Error("Failed to read the body of the profile for Auth0 token \""+token.AccessToken+"\":", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Unmarshall the JSON of the profile
	var profile map[string]interface{}
	if err := json.Unmarshal(raw, &profile); err != nil {
		commandMutex.Unlock()
		log.Error("Failed to unmarshall the profile:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Get the Auth0 user ID and username from the profile
	auth0ID := profile["user_id"].(string)
	auth0Username := profile["username"].(string)

	// Check to see if the requested person exists in the database
	var squelched int
	userID, username, admin, err := db.Users.Login(auth0ID)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if userID == 0 {
		// This is a new user, so add them to the database
		if userID, err = db.Users.Insert(auth0ID, auth0Username, ip); err != nil {
			commandMutex.Unlock()
			log.Error("Database error:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// By default, new users have the same stylizaition as their Auth0 username (all lowercase)
		username = auth0Username

		// By default, new users are not administrators
		admin = 0

		// By default, new users are not squelched
		squelched = 0

		// Log the user creation
		log.Info("Added \"" + username + "\" to the database (first login).")
	} else {
		// Check to see if this user is banned
		if userIsBanned, err := db.BannedUsers.Check(username); err != nil {
			commandMutex.Unlock()
			log.Error("Database error:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if userIsBanned == true {
			commandMutex.Unlock()
			log.Info("User \"" + username + "\" tried to log in, but they are banned.")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Check to see if this user is squelched
		if userIsSquelched, err := db.SquelchedUsers.Check(username); err != nil {
			commandMutex.Unlock()
			log.Error("Database error:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if userIsSquelched == true {
			squelched = 1
		} else {
			squelched = 0
		}

		// Update the database with last_login and last_ip
		if err := db.Users.SetLogin(username, ip); err != nil {
			commandMutex.Unlock()
			log.Error("Database error:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// Save the information to the session
	session.Values["userID"] = userID
	session.Values["username"] = username
	session.Values["admin"] = admin
	session.Values["squelched"] = squelched
	if err := session.Save(r, w); err != nil {
		commandMutex.Unlock()
		log.Error("Failed to save the session cookie:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Log the login request
	log.Info("User \""+username+"\" logged in from:", ip)

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()

	// Redirect them back to the main page
	http.Redirect(w, r, "/", http.StatusFound)
}

/*
 *  Logout function
 */

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Local variables
	functionName := "loginHandler"
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)

	// Lock the command mutex for the duration of the function to ensure synchronous execution
	commandMutex.Lock()

	// Check to see if their IP is banned
	if userIsBanned, err := db.BannedIPs.Check(ip); err != nil {
		commandMutex.Unlock()
		log.Error("Database error:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if userIsBanned == true {
		commandMutex.Unlock()
		log.Info("IP \"" + ip + "\" tried to logout, but they are banned.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get the session (this may be an empty session)
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		commandMutex.Unlock()
		log.Error("Unable to get the session during the", functionName, "function:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Delete all of their session variables
	delete(session.Values, "userID")
	delete(session.Values, "username")
	delete(session.Values, "admin")
	delete(session.Values, "squelched")
	if err := session.Save(r, w); err != nil {
		commandMutex.Unlock()
		log.Error("Failed to save the session cookie:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// The command is over, so unlock the command mutex
	commandMutex.Unlock()

	// Redirect them back to the main page
	http.Redirect(w, r, "/", http.StatusFound)
}
