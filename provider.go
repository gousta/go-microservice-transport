package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type responseEasysms struct {
	Status  string `json:"status"`
	Balance string `json:"balance"`
}

type responseNexmo struct {
	Status int8 `json:"status"`
}

type responseSMSAPI struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var requestClient = &http.Client{Timeout: 10 * time.Second}

func getRequest(u string, t interface{}) error {
	r, err := requestClient.Get(u)
	if err != nil {
		fmt.Println(err)
	}

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(t)
}

func postRequest(u string, v url.Values, t interface{}) error {
	r, err := requestClient.PostForm(u, v)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(t)
}

func easysms(t *Transaction) bool {
	u, _ := url.Parse("https://easysms.gr/api/sms/send")

	q := u.Query()
	q.Set("key", configuration.EasysmsKey)
	q.Set("type", "json")
	q.Set("from", t.Sender)
	q.Set("text", t.Message)
	q.Set("to", t.Receiver)
	u.RawQuery = q.Encode()

	response := responseEasysms{}
	t.Log("ATTEMPT WITH EASYSMS PROVIDER")
	getRequest(u.String(), &response)

	if response.Status != "1" {
		return false
	}

	return true
}

func nexmo(t *Transaction) bool {
	apiURL := "https://rest.nexmo.com/sms/json"

	data := url.Values{
		"api_key":    {configuration.NexmoKey},
		"api_secret": {configuration.NexmoSecret},
		"from":       {t.Sender},
		"text":       {t.Message},
		"to":         {t.Receiver},
		"type":       {"unicode"},
	}

	response := responseNexmo{}
	t.Log("ATTEMPT WITH NEXMO PROVIDER")
	postRequest(apiURL, data, &response)

	if response.Status != 0 {
		return false
	}

	return true
}

func smsapi(t *Transaction) bool {
	u, _ := url.Parse("https://api.smsapi.com/sms.do")

	q := u.Query()
	q.Set("access_token", configuration.SMSApiToken)
	q.Set("format", "json")
	q.Set("message", t.Message)
	q.Set("from", t.Sender)
	q.Set("to", t.Receiver)
	q.Set("encoding", "iso-8859-7")
	q.Set("datacoding", "gsm")
	u.RawQuery = q.Encode()

	response := responseSMSAPI{}
	t.Log("ATTEMPT WITH SMSAPI PROVIDER")
	getRequest(u.String(), &response)

	if response.Error != "" {
		return false
	}

	return true
}

func fakeOK(t *Transaction) bool {
	t.Log("ATTEMPT WITH FAKE PROVIDER")
	time.Sleep(3 * time.Second)
	return true
}

func fakeFail(t *Transaction) bool {
	return false
}

// TacticTest ...
func TacticTest(t *Transaction) {
	if fakeOK(t) {
		t.Sent()
	} else {
		t.Failed()
	}
}

// TacticTestFailing ...
func TacticTestFailing(t *Transaction) {
	if fakeFail(t) {
		t.Sent()
	} else {
		if t.Priority > 0 {
			if fakeFail(t) {
				t.Sent()
			} else {
				t.Failed()
			}
		} else {
			t.Failed()
		}
	}
}

// TacticMultiple ...
func TacticMultiple(t *Transaction) {
	if easysms(t) {
		t.Sent()
	} else {
		if t.Priority > 0 {
			if nexmo(t) {
				t.Sent()
			} else {
				t.Failed()
			}
		} else {
			t.Failed()
		}
	}
}

// TacticSingle ...
func TacticSingle(t *Transaction) {
	if nexmo(t) {
		t.Sent()
	} else {
		t.Failed()
	}
}
