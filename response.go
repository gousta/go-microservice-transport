package main

import (
	"encoding/json"
	"net/http"
)

type transactionResponse struct {
	Status      string      `json:"status"`
	Transaction Transaction `json:"transaction"`
}

type errorResponse struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func jsonResponse(w http.ResponseWriter) *json.Encoder {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w)
}
