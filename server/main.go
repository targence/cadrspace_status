package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func handleServer(conn net.Conn, requests chan []byte, responses chan []byte) {
	notify := make(chan error)
	defer conn.Close()

	go func() {
		buf := make([]byte, 1024)
		for {

			n, err := conn.Read(buf)
			if err != nil {
				notify <- err
				return
			}
			if n > 0 {
				req := make([]byte, n)
				copy(req, buf[:n])
				requests <- req
			}
		}
	}()

	for {
		select {
		case err := <-notify:
			if io.EOF == err {
				close(notify)
				return
			}
			log.Panic(err)
		case res := <-responses:
			conn.Write(res)
		}
	}

}

func handleTunnel(conn net.Conn, requests chan []byte, responses chan []byte) {
	defer conn.Close()
	notify := make(chan error)

	go func() {
		buf := make([]byte, 1024)
		for {

			n, err := conn.Read(buf)
			if err != nil {
				notify <- err
				return
			}
			if n > 0 {
				res := make([]byte, n)
				copy(res, buf[:n])
				responses <- res
			}
		}
	}()

	for {
		select {
		case err := <-notify:
			if io.EOF == err {
				close(notify)
				return
			}
			log.Panic(err)
		case r := <-requests:
			io.Copy(conn, bytes.NewBuffer(r))
		}
	}
}

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
		s.Response = "OK"
		s.UpdatedAt = time.Now().In(zone).Format(time.RFC3339)
		json, _ := json.Marshal(s)
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

const (
	statusAPIPort = "2000"
	tunnelPort    = "3000"
	serverPort    = "4000"
)

func main() {

	requests := make(chan []byte)
	responses := make(chan []byte)
	status := &status{}

	fmt.Println("Starting HTTP status api...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statusAPI(status, w, r)
	})
	go http.ListenAndServe(":"+statusAPIPort, nil)

	fmt.Println("Starting TCP server...")
	server, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Panic(err)
	}
	defer server.Close()

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				log.Panic(err)
			}
			go handleServer(conn, requests, responses)
		}
	}()

	fmt.Println("Starting TCP tunnel...")
	tunnel, err := net.Listen("tcp", ":"+tunnelPort)
	if err != nil {
		log.Panic(err)
	}
	defer tunnel.Close()

	for {
		conn, err := tunnel.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleTunnel(conn, requests, responses)
	}
}
