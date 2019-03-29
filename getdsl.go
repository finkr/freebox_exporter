package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// getDsl get dsl statistics
func getDsl() (int, int, int, int) {
	freeboxToken := os.Getenv("FREEBOX_TOKEN")
	if freeboxToken == "" {
		sessToken = getToken()
	}
	if sessToken == "" {
		sessToken = getSessToken(freeboxToken)
	}
	xdsl := database{
		DB:        "dsl",
		Fields:    []string{"rate_up", "rate_down", "snr_up", "snr_down"},
		Precision: 10,
		DateStart: int(time.Now().Unix() - 10),
	}
	client := http.Client{}
	r, err := json.Marshal(xdsl)
	if err != nil {
		log.Fatalln(err)
	}
	buf := bytes.NewReader(r)
	req, err := http.NewRequest("POST", fmt.Sprintf("%sapi/%s/rrd/", mafreebox, version), buf)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("X-Fbx-App-Auth", sessToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 404 {
		log.Fatal(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(body, &rrdTest)
	switch rrdTest.ErrorCode {
	case "auth_required":
		sessToken = getSessToken(freeboxToken)
	case "invalid_token":
		log.Fatalln("The app token you are trying to use is invalid or has been revoked")
	case "pending_token":
		log.Println("The app token you are trying to use has not been validated by user yet")
	case "insufficient_rights":
		log.Fatalln("Your app permissions does not allow accessing this API")
	case "denied_from_external_ip":
		log.Fatalln("You are trying to get an app_token from a remote IP")
	case "invalid_request":
		log.Fatalln("Your request is invalid")
	case "ratelimited":
		log.Fatalln("Too many auth error have been made from your IP")
	case "new_apps_denied":
		log.Fatalln("New application token request has been disabled")
	case "apps_denied":
		log.Fatalln("API access from apps has been disabled")
	case "internal_error":
		log.Fatalln("Internal error")
	}
	if len(rrdTest.Result.Data) > 0 {
		return rrdTest.Result.Data[0]["rate_up"], rrdTest.Result.Data[0]["rate_down"], rrdTest.Result.Data[0]["snr_up"], rrdTest.Result.Data[0]["snr_down"]
	} else {
		return 0,0,0,0
	}
}
