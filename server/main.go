package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type status struct {
	Error     string `json:"error,omitempty"`
	Response  string `json:"response"`
	UpdatedAt string `json:"updated_at"`
}

type statusRes struct {
	API   string `json:"api"`
	Cache struct {
		Schedule string `json:"schedule"`
	} `json:"cache"`
	Cam     []string `json:"cam"`
	Contact struct {
		Irc       string `json:"irc"`
		IssueMail string `json:"issue_mail"`
		Ml        string `json:"ml"`
		Twitter   string `json:"twitter"`
	} `json:"contact"`
	IssueReportChannels []string `json:"issue_report_channels"`
	Location            struct {
		Address string  `json:"address"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	} `json:"location"`
	Logo     string   `json:"logo"`
	Projects []string `json:"projects"`
	Space    string   `json:"space"`
	Spacefed struct {
		Spacenet   bool `json:"spacenet"`
		Spacephone bool `json:"spacephone"`
		Spacesaml  bool `json:"spacesaml"`
	} `json:"spacefed"`
	State struct {
		Open bool `json:"open"`
	} `json:"state"`
	URL string `json:"url"`
}

var statusTemplate = `{
    "api": "0.13",
    "space": "CADR",
    "logo": "http://cadrspace.ru/w/skins/common/images/cadr.png",
    "url": "http://cadrspace.ru",
    "location": {
        "address": "Nizhniy Novgorod, Russian Federation, Studentcheskaya st. 6, aud. 054",
        "lon": 43.988235,
        "lat": 56.302663
    },
    "spacefed": {
        "spacenet": false,
        "spacesaml": false,
        "spacephone": false
    },
    "contact": {
        "twitter": "@cadrspace",
        "irc": "irc://chat.freenode.net/##cadr",
        "ml": "cadr-hackerspace@googlegroups.com",
        "issue_mail": "poptsov.artyom@gmail.com"
    },
    "cam": [
        "http://nntc.nnov.ru:58080/?action=stream"
    ],
    "issue_report_channels": [
        "issue_mail",
        "ml"
    ],
    "state": {
        "open": false
    },
    "projects": [
        "https://github.com/cadrspace"
    ],
    "cache": {
        "schedule": "m.05"
    }
}
`

const (
	statusAPIPort = "2000"
)

var zone, _ = time.LoadLocation("Europe/Moscow")

func statusAPI(s *status, w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		data := statusRes{}
		jsonErr := json.Unmarshal([]byte(statusTemplate), &data)
		if jsonErr != nil {
			log.Println(jsonErr)
		}

		if s.UpdatedAt != "" {
			t0, err := time.Parse(
				time.RFC3339,
				s.UpdatedAt)

			if err != nil {
				log.Println(err)
			}

			t1 := time.Now()

			duration := t1.Sub(t0)
			if duration.Seconds() < (70 * time.Second).Seconds() {
				data.State.Open = true
			}
		}

		json, _ := json.Marshal(data)
		w.WriteHeader(http.StatusOK)
		w.Write(json)

	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		values, err := url.ParseQuery(string(body))
		if err != nil {
			log.Println(err)
		}

		if values.Get("open") == "true" {

			s.Response = "OK"
			s.UpdatedAt = time.Now().In(zone).Format(time.RFC3339)
			json, err := json.Marshal(s)
			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusOK)
			w.Write(json)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func main() {
	status := &status{}

	fmt.Println("Starting HTTP status api...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statusAPI(status, w, r)
	})
	http.ListenAndServe(":"+statusAPIPort, nil)
}
