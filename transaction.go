package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Transaction Struct
type Transaction struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	Status    string        `json:"status"`
	Receiver  string        `json:"receiver"`
	Message   string        `json:"message"`
	Sender    string        `json:"sender"`
	Tags      string        `json:"tags"`
	Attempts  int8          `json:"attempts"`
	Priority  int8          `json:"priority"`
	Timeout   int64         `json:"timeout,omitempty" bson:"timeout"`
	CreatedAt int64         `json:"created_at,omitempty" bson:"created_at"`
}

// Init setups some initial variables for a Transaction
func (t *Transaction) Init() {
	t.ID = bson.NewObjectId()
	t.CreatedAt = time.Now().Unix()
	t.Status = "received"
	t.Attempts = 0
}

// LogRequest ...
func (t *Transaction) LogRequest() {
	t.Log("RECEIVED")
}

// Validate ...
func (t *Transaction) Validate() (valid bool, errorCode string, reason string) {
	if t.Message == "" {
		return false, "required:message", "`message` is required"
	}
	if t.Sender == "" {
		return false, "required:sender", "`sender` is required"
	}

	valid, err := regexp.MatchString(`^\+[1-9]{1}[0-9]{3,14}$`, t.Receiver)
	if err != nil {
		log.Fatal(err)
		return false, "regex:fail", "regex failed"
	}
	if !valid {
		return false, "regex:invalid-format", "`receiver` format is invalid - format: +3069XXXXXXXX"
	}

	return valid, errorCode, reason
}

// IsUnique checks if similar (Transaction Receiver/Message/Sender) exists
func (t *Transaction) IsUnique() (unique bool, duplicate Transaction) {
	result := Transaction{}

	query := bson.M{
		"receiver": t.Receiver,
		"message":  t.Message,
		"sender":   t.Sender,
		"tags":     t.Tags,
	}

	Transactions.Find(query).One(&result)

	if result.ID != "" {
		t.Log("DUPLICATE")
		return false, result
	}

	return true, Transaction{}
}

// Save inserts a Transaction in the DB
func (t *Transaction) Save() bool {
	err := Transactions.Insert(t)
	if err != nil {
		log.Println("[MONGO] ERROR SAVING TRANSACTION:", err.Error())
	}

	t.Log("ACCEPTED")
	return true
}

// FindByID Finds record by ID
func (t *Transaction) FindByID(searchID string) {
	if bson.IsObjectIdHex(searchID) {
		err := Transactions.FindId(bson.ObjectIdHex(searchID)).One(&t)
		if err != nil {
			log.Println("[MONGO] ERROR FINDING TRANSACTION:", err.Error())
		}
	}
}

// Sent marks transaction as Sent
func (t *Transaction) Sent() {
	err := Transactions.UpdateId(t.ID, bson.M{"$set": bson.M{"status": "sent"}})
	if err != nil {
		log.Println("[MONGO] ERROR SET STATUS AS SENT:", err.Error())
	}
	t.Log("SENT")
}

// Failed marks transaction as Failed
func (t *Transaction) Failed() {
	err := Transactions.UpdateId(t.ID, bson.M{"$set": bson.M{"status": "failed"}})
	if err != nil {
		log.Println("[MONGO] ERROR SETTING STATUS AS FAILED:", err.Error())
	}
	t.Log("FAILED")
}

// Send ...
func (t Transaction) Send() {
	t.Log("QUEUED")

	err := Transactions.UpdateId(t.ID, bson.M{"$inc": bson.M{"attempts": 1}})
	if err != nil {
		log.Println("[MONGO] ERROR ON INCREMENTING ATTEMPT:", err.Error())
	}

	if strings.Contains(t.Tags, "test") {
		TacticTest(&t)
	} else {
		TacticSingle(&t)
	}
}

// Log ...
func (t *Transaction) Log(text string) {
	log.Printf("[TRANSACTION] %s [%s_%d_%s_%s_%s_%s]", text, t.ID.Hex(), t.Priority, t.Sender, t.Receiver, t.Message, t.Tags)
}
