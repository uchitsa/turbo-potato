package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type transport struct {
	dialer    *net.Dialer
	rt        http.RoundTripper
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

func main() {
	file, err := os.Open("sites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	res := make(map[string]bool)

	ntp := NewTransport()
	_ = &http.Client{Transport: ntp}

	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}
		line = cutEndOfLine(line)
		res[line] = ping(line)
	}

	for k, v := range res {
		log.Printf("site: %s available: %v\n", k, v)
	}
}

func cutEndOfLine(line string) string {
	return line[:len(line)-1]
}

func ping(url string) bool {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", url), timeout)
	if err != nil {
		log.Println("site unreachable error: ", err)
		return false
	}

	defer conn.Close()

	return true
}

func NewTransport() *transport {
	return &transport{
		dialer: &net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
		},
		rt: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.reqStart = time.Now()
	response, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	t.reqEnd = time.Now()

	return response, nil
}
