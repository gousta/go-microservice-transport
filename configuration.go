package main

// Configuration ...
type Configuration struct {
	Port       string
	MongoDBUrl string

	PapertrailHost string
	PapertrailPort int

	TransactionRetries int
	TransactionQueue   int

	EasysmsKey  string
	NexmoKey    string
	NexmoSecret string
	SMSApiToken string
}
