package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: time.Second * 5,
}

type status struct {
	Error    string `json:"error,omitempty"`
	Response string `json:"response"`
}

const (
	statusHTTPServer = "http://localhost:2000"
)

func pushStatus() {

	for c := time.Tick(30 * time.Second); ; <-c { // instant start Ticker

		payload := strings.NewReader("open=true")
		req, err := http.NewRequest("POST", statusHTTPServer, payload)
		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Set("User-Agent", "cadr-status")
		req.PostForm = url.Values{"open": {"open"}}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Output URL:
		// fmt.Println(req.URL.String())

		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}

		defer resp.Body.Close()

		data := status{}
		jsonErr := json.Unmarshal(body, &data)
		if jsonErr != nil {
			log.Println(jsonErr)
		}

		if data.Error != "" {
			log.Println(data.Error)
		}

		if resp.StatusCode != 200 {
			log.Println(err, resp.Status)
		}
	}
}

func main() {

	fmt.Println("Starting Cadrspace client...")
	pushStatus()

}
