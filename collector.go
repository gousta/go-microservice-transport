package main

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const queuedItemTimeoutSeconds = 60 // 3 minutes

// StartCollector ...
func StartCollector() {
	log.Println("[COLLECTOR] START")

	regularTicker := time.NewTicker(time.Second * 2)
	otherTicker := time.NewTicker(time.Second * 4)

	go func() {
		for range regularTicker.C {
			regularCollector(configuration.TransactionQueue)
		}
	}()

	go func() {
		for range otherTicker.C {
			failedCollector(configuration.TransactionRetries)
			unfinishedCollector()
		}
	}()
}

func regularCollector(queue int) {
	var pending = bson.M{"status": "received"}
	var queueStatus = mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"status":  "queued",
				"timeout": time.Now().Unix() + queuedItemTimeoutSeconds,
			},
		},
		ReturnNew: true,
	}

	findings, err := Transactions.Find(pending).Count()
	if err != nil {
		log.Println("[MONGO] ERROR COLLECTING RECEIVED:", err.Error())
	}

	if findings > 0 {
		max := findings

		if findings > queue {
			max = queue
		}

		for j := 0; j <= max; j++ {
			go func() {
				t := Transaction{}

				Transactions.Find(pending).Apply(queueStatus, &t)

				if t.ID != "" {
					t.Send()
				}
			}()
		}
	}
}

func failedCollector(retries int) {
	var query = bson.M{
		"status": "failed",
		"attempts": bson.M{
			"$lt": retries,
		},
		"priority": bson.M{
			"$gt": 0,
		},
	}

	var update = bson.M{
		"$set": bson.M{
			"status": "received",
		},
	}

	_, err := Transactions.UpdateAll(query, update)
	if err != nil {
		log.Println("[MONGO] ERROR FETCHING FAILED TRANSACTIONS:", err.Error())
	}
}

func unfinishedCollector() {
	var query = bson.M{
		"status": "queued",
		"timeout_at": bson.M{
			"$lt": time.Now().Unix(),
		},
	}

	var update = bson.M{
		"$set": bson.M{
			"status": "received",
		},
	}

	_, err := Transactions.UpdateAll(query, update)
	if err != nil {
		log.Println("[MONGO] ERROR FETCHING UNFINISHED TRANSACTIONS:", err.Error())
	}
}
