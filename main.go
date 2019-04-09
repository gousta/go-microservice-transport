package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	static        = http.StripPrefix("/", http.FileServer(http.Dir("./ui/build")))
	configuration = Configuration{}

	// Transactions ...
	Transactions *mgo.Collection
)

func init() {
	log.Println("[APP] BOOT")

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	configuration.Port = os.Getenv("PORT")
	configuration.MongoDBUrl = os.Getenv("MONGODB_URI")

	configuration.PapertrailHost = os.Getenv("PAPERTRAIL_HOST")
	configuration.PapertrailPort, _ = strconv.Atoi(os.Getenv("PAPERTRAIL_PORT"))

	configuration.TransactionRetries, _ = strconv.Atoi(os.Getenv("TRANSACTION_RETRIES"))
	configuration.TransactionQueue, _ = strconv.Atoi(os.Getenv("TRANSACTION_QUEUE"))

	configuration.EasysmsKey = os.Getenv("EASYSMS_KEY")
	configuration.NexmoKey = os.Getenv("NEXMO_KEY")
	configuration.NexmoSecret = os.Getenv("NEXMO_SECRET")
	configuration.SMSApiToken = os.Getenv("SMSAPI_TOKEN")

	mgosession, err := mgo.Dial(configuration.MongoDBUrl)

	if err != nil {
		log.Println("[MONGO] ERROR CONNECTING:", err.Error())
		os.Exit(1)
	} else {
		log.Println("[MONGO] CONNECTED")
	}

	// Transactions ...
	Transactions = mgosession.DB("").C("transactions")
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func newTransaction(w http.ResponseWriter, r *http.Request) {
	// Build Transaction
	t := Transaction{}
	_ = json.NewDecoder(r.Body).Decode(&t)

	// Set initial values
	t.Init()

	// Log
	t.LogRequest()

	// Validate
	valid, code, reason := t.Validate()
	if !valid {
		jsonResponse(w).Encode(errorResponse{
			Status:  "error",
			Code:    code,
			Message: "Failed to validate: " + string(reason),
		})
		return
	}

	// Check Uniqueness
	unique, dt := t.IsUnique()
	if unique {
		// Persist Transaction in MongoDB
		t.Save()

		jsonResponse(w).Encode(transactionResponse{
			Status:      "ok",
			Transaction: t,
		})
	} else {
		jsonResponse(w).Encode(transactionResponse{
			Status:      "ok",
			Transaction: dt,
		})
	}
}

func getTransaction(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	t := Transaction{}
	t.FindByID(params["id"])

	if t.ID == "" {
		log.Printf("[TRANSACTION] 404 (%s)", params["id"])
		jsonResponse(w).Encode(errorResponse{
			Status:  "error",
			Message: "Transaction does not exist",
		})
	} else {
		t.Log("GET")
		jsonResponse(w).Encode(t)
	}
}

func getPort() string {
	return ":" + configuration.Port
}

func main() {
	log.Println("[APP] START")

	StartCollector()

	rt := mux.NewRouter().StrictSlash(true)

	rt.HandleFunc("/v1/ping", ping).Methods("GET")
	rt.HandleFunc("/v1/transaction", newTransaction).Methods("POST")
	rt.HandleFunc("/v1/transaction/{id}", getTransaction).Methods("GET")
	rt.PathPrefix("/").Handler(static)

	// these two lines are important in order to allow access from the front-end side to the methods
	origins := handlers.AllowedOrigins([]string{"*"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"})

	// launch server with CORS validations
	err := http.ListenAndServe(getPort(), handlers.CORS(origins, methods)(rt))
	if err != nil {
		log.Println("[SERVER] STARTED")
	} else {
		log.Println("[SERVER] FAILED TO START", err.Error())
	}
}
