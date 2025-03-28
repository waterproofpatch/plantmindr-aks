// entry point to golang server.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"app/app"

	"github.com/waterproofpatch/go_authentication/authentication"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func makeRouter() *mux.Router {
	router := mux.NewRouter()

	return router
}

var DEFAULT_PORT = 8080

// startServing creates the server mux and registers endpoints with it.
func startServing(port int, router *mux.Router) {
	portStr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Starting server on https://%s...", portStr)

	methods := []string{"GET", "POST", "PUT", "DELETE"}
	headers := []string{"Content-Type", "Access-Control-Allow-Origin", "Authorization"}
	origins := []string{
		// from the storage account
		"https://plantmindrstorage.z13.web.core.windows.net",
		"https://www.plantmindr.com",
		"https://plantmindr.com",
		"www.plantmindr.com",
		"plantmindr.com",
		"http://localhost:4200",
		"https://localhost:4200",
		"https://antlion.azurewebsites.net",
		"http://antlion.azurewebsites.net",
	}
	srv := &http.Server{
		// Handler: router,
		Handler: handlers.CORS(handlers.AllowCredentials(),
			handlers.AllowedMethods(methods),
			handlers.AllowedHeaders(headers),
			handlers.AllowedOrigins(origins))(router),
		Addr: portStr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// main is the entrypoint to the program.
func main() {
	log.Printf("Starting...")
	stopCh := make(chan bool)

	// Stop the function by sending a message to the channel
	// stopCh <- true

	router := makeRouter()
	dropTables := false
	port := DEFAULT_PORT
	if os.Getenv("DROP_TABLES") == "true" {
		fmt.Printf("DROP_TABLES=true")
		dropTables = true
	} else {
		fmt.Printf("DROP_TABLES=false")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Printf("Error converting port %s to int.", os.Getenv("PORT"))
		return
	}
	log.Printf("Port will be %d", port)
	log.Printf("Default admin user name will be %s", os.Getenv("DEFAULT_ADMIN_USERNAME"))

	// must happen before we get the db
	registrationCallbackUrl := "https://www.plantmindr.com/authentication?mode=login&verified=true"
	var dbUrl string = ""
	var secret string = ""
	var refreshSecret string = ""
	var defaultAdminEmail string = ""
	var defaultAdminUsername string = ""
	var defaultAdminPassword string = ""
	if os.Getenv("DEBUG") == "true" {
		registrationCallbackUrl = "https://localhost:4200/authentication?mode=login&verified=true"
		fmt.Println("Using env DATABASE_URL")
		dbUrl = os.Getenv("DATABASE_URL")
		secret = os.Getenv("SECRET")
		refreshSecret = os.Getenv("SECRET")
		defaultAdminEmail = os.Getenv("DEFAULT_ADMIN_EMAIL")
		defaultAdminUsername = os.Getenv("DEFAULT_ADMIN_USERNAME")
		defaultAdminPassword = os.Getenv("DEFAULT_ADMIN_PASSWORD")
	} else {
		dbUrl, err = app.GetSecret("sqlDbPassword", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
		secret, err = app.GetSecret("secret", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
		refreshSecret, err = app.GetSecret("refreshSecret", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
		defaultAdminEmail, err = app.GetSecret("defaultAdminEmail", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
		defaultAdminUsername, err = app.GetSecret("defaultAdminUsername", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
		defaultAdminPassword, err = app.GetSecret("defaultAdminPassword", "plantmindrrbackv")
		if err != nil {
			panic("Error getting secret")
		}
	}

	authentication.Init(
		secret,
		refreshSecret,
		defaultAdminEmail,
		defaultAdminUsername,
		defaultAdminPassword,
		router,
		dbUrl,
		dropTables,
		true, // requiresVerificaiton
		app.ResetPasswordCallback,
		app.RegistrationCallback,
		registrationCallbackUrl,
		os.Getenv("DEBUG") == "true")

	db := authentication.GetDb()

	app.InitViews(router)
	app.InitModels(db, dropTables)

	// Run the function in a goroutine
	go app.StartTimer(stopCh, db)

	startServing(port, router)
}
