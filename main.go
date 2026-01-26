package main

// want to write a print stufff in the terminal
import (
	"backend/config"
	"backend/database"
	"backend/handlers"
	"backend/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// main go file has three important componets, the Gorilla mux routers, the pgx postgresl driver, and Goth for atuh

// the os var will be used to read enviorment var
// make struct for User type of your contact page
func main() {
	// lets check if the
	// normall i would write, load, err but the godotenv
	err := godotenv.Load("go.env")
	if err != nil {
		log.Fatal("No env file was found")
	}
	// get the gotenv file
	//addres := os.Getenv("DB_HOST, DP_PORT,DB_USER, DB_NAME")
	//if addres == "" {
	//	log.Fatal("Did not find the db")
	//}
	// here we passing a func called connectionstring, the connection string will have all the env vars
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_NAME", "postgres"),
	)
	// here use the log.Print
	log.Print("Connecting to database")

	// here we will actually initalize the actual db
	// in the doucmentaton is pgx.Connect(context.Background(), os.Getenv("DATABASE_URL")), instead of wirting each dbport, name, etc in the parameter
	// package context which carries deadlines, cancellation signals and will cancel when needed it
	dbConn, err := database.Newinit(context.Background(), connString)
	if err != nil {
		log.Fatalf("Could not connect to database%v", err)
	}
	// now we close the connecton using the defer close
	defer dbConn.Close()

	// test connection and check if we can even co
	if err := dbConn.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping databse")
	}
	log.Println("Databse connection successful")
	// here we create the tables

	if err := dbConn.CreateTables(context.Background()); err != nil {
		log.Fatal("Failed to create tabels", err)
	}
	log.Println("Databse tables are ready to go")

	// here ill refer to OAuth config file here
	// call the name of the file and function that comes with
	config.InitAuth()
	//config.GetSessionStore()
	log.Print("OAuth is ready to go")
	// for this instance i am going to make router here, for future use ill put the routes in the routes folder

	// so here we will start created the new router
	router := mux.NewRouter()
	// all the authhandlers are reffered through dot notation
	AuthHandler := handlers.NewAuthHandler(dbConn)
	// the setupRoutes(routes reffers to the mux router, then the handler)
	setupRoutes(router, AuthHandler)
	// Middlewares can be added to a router using Router.Use():
	// follow this strucutre routes.Use(name of file.methodname)
	//routes.Use(middleware.LoggingMiddleware)
	//routes.Use(middleware.AuthMiddleware)
	// routes.Use(middleware.CorsMiddleware) add this if you dont want the corshandler var
	// here ill chain them, if you dont want to use this then you can use the routes.Use(add middleware to routes)
	Corshandler := middleware.LoggingMiddleware(
		middleware.CorsMiddleware(
			// dont use middleware.authmiddleware(router) this means we apply a auth route globally before anyone signs in
			// when we get
			router,
		),
	)
	log.Printf("Middleware works!")
	// to create a graceful shutdown we specifcy how long we want the server to run
	srv := &http.Server{
		Addr: ":" + getEnv("PORT", "8080"), // this would be the port we will run on the backend. Note addr == adress
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      Corshandler, // Pass our instance of gorilla/mux in.
	}
	// / Run our server in a goroutine so that it doesn't block.
	// this is take straight from the doucmentation from the documentation from gorilla mux
	go func() {
		log.Printf("Server starting on http://localhost:%s", getEnv("PORT", "8080"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	//
	// this means control c to stop for the server to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// block until we receive our signal
	<-c
	// create a deadline to wait for. Most of this is from the doucmentation
	// here we will need to pass three arguments, context, time, and err
	ctx, cancel := context.WithTimeoutCause(context.Background(), 10*time.Second, err)
	defer cancel()
	srv.Shutdown(ctx)

	// cleanup expried tokens and sessions
	dbConn.DeleteExpiredPasswordResetTokens((context.Background()))
	dbConn.DeleteExpiredSessions(context.Background())
	dbConn.Close()
}

// create a subrouter function
func setupRoutes(router *mux.Router, authHandler *handlers.AuthHandler) {
	// API prefix
	api := router.PathPrefix("/api").Subrouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","service":"auth-api"}`)
	}).Methods("GET")

	// Auth routes - Registration & Login
	api.HandleFunc("/auth/register", authHandler.RegisterHandler).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.LoginHandler).Methods("POST")

	// OAuth routes
	api.HandleFunc("/auth/{provider}", authHandler.BeginAuthHandler).Methods("GET")
	api.HandleFunc("/auth/{provider}/callback", authHandler.CallbackHandler).Methods("GET")

	// Password reset routes
	api.HandleFunc("/auth/forgot-password", authHandler.ForgotPasswordHandler).Methods("POST")
	api.HandleFunc("/auth/reset-password", authHandler.ResetPasswordHandler).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/auth/me", authHandler.GetCurrentUserHandler).Methods("GET")
	protected.HandleFunc("/auth/logout", authHandler.LogoutHandler).Methods("POST")
	protected.HandleFunc("/auth/change-password", authHandler.ChangePasswordHandler).Methods("POST")

	log.Println(" Routes configured")
}

// have this serve as the fallback
// which means if one of the
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// this is basic overview of a handler
// The http.ResponseWriter is used to construct the HTTP response,
// while the *http.Request contains information about the incoming HTTP request.
// this will be used get a response after someone submits the contactform and fills in their requirments
/*type ContactForm struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

// we can update the email and phone etc onc we upload/change the ui
func getContact(w http.ResponseWriter, r *http.Request) {
	// in your contact form we want acces to our items, this will be set in a dict
	//Interfaces can be used as both keys and values in maps, offering flexibility.
	//Interfaces as Map Values: This is a common pattern, especially map[string]interface{}, which allows you to store values of different concrete types under string keys. You will need to use type assertions or type switches to work with the underlying concrete types.
	// contanctinfo := map[string] interface{}  or you can use the make method to make a new dict
	contactInformation := make(map[string]interface{})
	contactInformation["email"] = "support@citizenshipprep.com"
	contactInformation["phone"] = 1 - 800 - 24456 // might swap this to a string
	contactInformation["address"] = "123 Learning Ave, San Francisco, CA 94102"
	contactInformation["hours"] = "Monday-Friday: 9AM-6PM EST"
	// this tells the client we are sending a json data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contactInformation)
}

// you need to create func for every handlefunc in the routes
// this will happend after they submit the contact form
func submitContactForm(w http.ResponseWriter, r *http.Request) {
	// use the struct for the users
	// here we need to handle different kinds of logic, such as did they submit the
	// here we reffer to the struct
	var userContactForm ContactForm
	// this takes one argument
	err := json.NewDecoder(r.Body).Decode(&userContactForm)
	// we check if there is a err
	// anytime we want a  http request we use the w,r, w== send request back to client, r== tell us what the request
	if err != nil {
		log.Print("Error Handling Json")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// here we will check if the contact form does not work/ or is not filled in the firstname
	if userContactForm.FirstName == "" {
		log.Println("Fill in the First name section")
		// http error send it back to the client and give the them an error
		http.Error(w, "First name is required", http.StatusBadRequest)
	}
	// now we check if the lastname is section
	if userContactForm.LastName == "" {
		log.Println("Fill in the Last name section")
		http.Error(w, "Last name is required", http.StatusBadRequest)
		return
	}
	if userContactForm.Email == "" {
		log.Println("Fill in the Email section")
		http.Error(w, "Emai is required", http.StatusBadRequest)
		return
	}
	if userContactForm.Message == "" {
		log.Println("Fill in the Message section")
		http.Error(w, "Fill in the message", http.StatusBadRequest)
		return
	}
	//

	// make another var to say thank you for the response
	// Log the received form data
	log.Println("📧 Contact Form Data:")
	log.Printf("   👤 Name: %s %s", userContactForm.FirstName, userContactForm.LastName)
	log.Printf("   📧 Email: %s", userContactForm.Email)
	log.Printf("   📋 Subject: %s", userContactForm.Subject)
	log.Printf("   💬 Message: %s", userContactForm.Message)
	log.Println("========================================")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"message": "Thank you for your message. We'll get back to you soon!",
		"status":  "success",
	}

	json.NewEncoder(w).Encode(response)
	log.Println("✅ Success response sent to frontend")

}

func main() {
	//http.HandleFunc("/", handler) // the Handlefunc takes the func and sets a route to it
	//log.Fatal(http.ListenAndServe(":8080", nil)) // the listen an serve takes the host

	//http.handleContactfunc() you can only use the http.HandleFunc since this is a method reffering to a object

	// so here we will start the newRouter
	routes := mux.NewRouter()
	// Middlewares can be added to a router using Router.Use():
	routes.Use(LoggingMiddleware)

	// here we will make a contact route and we will make sure the contact page
	//routes.HandleFunc("/Contact", Contact)
	//routes.HandleFunc("/About", About)
	//routes.HandleFunc("Payment", Payment)
	// just like the http.HandleFunc, we have one for mux called http.handle("/",routes)
	//http.Handle("/", routes)
	//http.ListenAndServe("8080",nil)

	// this if we have mutiple routes to one address for your contact page
	api := routes.PathPrefix("/api").Subrouter()
	api.HandleFunc("/Contact", getContact).Methods("GET")
	api.HandleFunc("/Contact/submit", submitContactForm).Methods("POST")

	// starts the server
	log.Println("The server is about to start")
	log.Fatal(http.ListenAndServe(":8080", routes))
}

// here we make cors for frontend communcation.
// will handle if the contact page is submited and ready
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set the cors handler
		//w.Header().Set("Access-Control-Allow-Origin", "*") // access control for requests with http.Only
		// * this means anyone from any website can talk

		// Production (more secure!) this means only our local host can talk so
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // we can add put and patch for the other files that require more complex routes
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// checks if the requests work and check if the preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)

	})

	/*db, err := pgxpool.New(ctx, connstring)
		if err != nil {
			 return fmt.Errorf("unable to parse connection string: %w", err)
		}
		//Step 2:  test if there is connection
		if err := db.Ping(ctx); err != nil{
			return fmt.Errorf("unable to ping database: %w", err)
		}
	 // Step 3: close the database at the end
		defer db.Close()

		return nil
	} */
