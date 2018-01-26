package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/codeship/go-retro"
)

type Client struct {
}

func (c *Client) NewTunnel() *Tunnel {
	return &Tunnel{C: c}
}

type Tunnel struct {
	C       *Client
	CloseCh chan struct{}
}

type conn struct {
	T          *Tunnel
	RemoteConn net.Conn
	LocalConn  net.Conn
}

func (t *Tunnel) Close() {
	close(t.CloseCh)
}

func (t *Tunnel) Closing() <-chan struct{} {
	return t.CloseCh
}

var (
	ErrNetwork = retro.NewStaticRetryableError(errors.New("error: Max retries attempts reached"), 10000, 2)
)

func (c *conn) open() {
	var err error

	err = retro.DoWithRetry(func() error {
		c.RemoteConn, err = net.Dial("tcp", tunnelServer)
		if err != nil {
			log.Println(err)
			return ErrNetwork
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	err = retro.DoWithRetry(func() error {
		c.LocalConn, err = net.Dial("tcp", localPortForward)
		if err != nil {
			log.Println(err)
			return ErrNetwork
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	c.pipe()
}

func (c *conn) close() {
	if c.LocalConn != nil {
		c.LocalConn.Close()
	}

	if c.RemoteConn != nil {
		c.RemoteConn.Close()
	}
}

func (c *conn) pipe() {
	errorCh := make(chan error)
	remoteCh := chanFromConn(c.RemoteConn, errorCh)
	localCh := chanFromConn(c.LocalConn, errorCh)

	for {
		select {
		case b := <-remoteCh:
			c.LocalConn.Write(b)
		case b := <-localCh:
			c.RemoteConn.Write(b)
		case <-errorCh:
			c.close()
			c.open()
			return
		case <-c.T.CloseCh:
			c.close()
			return
		}
	}
}

func chanFromConn(conn net.Conn, errorCh chan error) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				errorCh <- err
				break
			}
		}
	}()

	return c
}

var client = &http.Client{
	Timeout: time.Second * 5,
}

type status struct {
	Error    string `json:"error,omitempty"`
	Response string `json:"response"`
}

func pushStatus() {

	for c := time.Tick(30 * time.Second); ; <-c { // instant start Ticker

		req, err := http.NewRequest("POST", statusHTTPServer, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Set("User-Agent", "cadr-status")
		q := req.URL.Query()
		q.Add("open", "true")

		req.URL.RawQuery = q.Encode()

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

const (
	statusHTTPServer = "http://localhost:2000"
	tunnelServer     = "localhost:3000"
	localPortForward = "5000"
)

func main() {
	c := &Client{}
	t := c.NewTunnel()

	t.CloseCh = make(chan struct{})

	connection := &conn{T: t}
	fmt.Println("Starting TCP client...")
	go connection.open()
	go pushStatus()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	func() {
		for s := range sig {
			fmt.Printf("%v received\n", s)
			t.Close()
			break
		}
	}()

	<-t.Closing()
	fmt.Println("Bye! tunnel closed")
}
